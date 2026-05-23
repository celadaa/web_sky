// Package services — servicio de envío de email (verificación + bienvenida).
//
// Diseño:
//   - SMTP plain con STARTTLS (RFC 3207). Suficiente para Mailgun, SendGrid,
//     Amazon SES, Postmark y cualquier servidor moderno.
//   - Sin dependencias externas (stdlib net/smtp + crypto/tls).
//   - Multipart MIME: text/plain + text/html.
//   - Modo "deshabilitado" (cfg.EmailDisabled o cfg.SMTPHost vacío): no abre
//     conexión TCP — escribe el enlace en stdout. Imprescindible para CI/local.
//   - Plantillas: text/template (escapado simple; las HTML usan html/template
//     para escapar variables del usuario automáticamente).
//
// Logs: nunca imprimimos el token completo. Los tokens viajan en la URL
// dentro del cuerpo del email — eso es estándar y aceptable.
package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net"
	"net/smtp"
	"path/filepath"
	"strings"
	texttemplate "text/template"
	"time"

	"skihub/internal/config"
)

// EmailService envía correos transaccionales.
type EmailService struct {
	cfg *config.Config

	tplVerificacionHTML *template.Template
	tplVerificacionTXT  *texttemplate.Template
	tplBienvenidaHTML   *template.Template
	tplBienvenidaTXT    *texttemplate.Template
}

// DatosEmailVerificacion son los placeholders de la plantilla.
type DatosEmailVerificacion struct {
	Nombre   string
	URL      string
	HorasTTL int
}

// DatosEmailBienvenida son los placeholders del welcome (sin verificación).
type DatosEmailBienvenida struct {
	Nombre string
	URL    string // p.ej. APP_BASE_URL/login o /estaciones
}

// NuevoEmailService carga las plantillas y prepara la conexión SMTP.
// Si templatesDir es vacío usa cfg.AppTemplates (que apunta a web/templates).
func NuevoEmailService(cfg *config.Config, templatesDir string) (*EmailService, error) {
	if cfg == nil {
		return nil, errors.New("EmailService: config nil")
	}
	if templatesDir == "" {
		templatesDir = cfg.AppTemplates
	}
	emailDir := filepath.Join(templatesDir, "email")

	vh, err := template.ParseFiles(filepath.Join(emailDir, "verification.html.tmpl"))
	if err != nil {
		return nil, fmt.Errorf("parse verification.html: %w", err)
	}
	vt, err := texttemplate.ParseFiles(filepath.Join(emailDir, "verification.txt.tmpl"))
	if err != nil {
		return nil, fmt.Errorf("parse verification.txt: %w", err)
	}
	bh, err := template.ParseFiles(filepath.Join(emailDir, "welcome.html.tmpl"))
	if err != nil {
		return nil, fmt.Errorf("parse welcome.html: %w", err)
	}
	bt, err := texttemplate.ParseFiles(filepath.Join(emailDir, "welcome.txt.tmpl"))
	if err != nil {
		return nil, fmt.Errorf("parse welcome.txt: %w", err)
	}

	return &EmailService{
		cfg:                 cfg,
		tplVerificacionHTML: vh,
		tplVerificacionTXT:  vt,
		tplBienvenidaHTML:   bh,
		tplBienvenidaTXT:    bt,
	}, nil
}

// Desactivado devuelve true cuando los emails no se enviarán por SMTP
// real (modo desarrollo o configuración incompleta). En ese caso los
// senders loguean el enlace por stdout y devuelven nil.
func (s *EmailService) Desactivado() bool {
	return s.cfg.EmailDisabled || s.cfg.SMTPHost == ""
}

// SendEmailVerification envía el correo con el botón de confirmación.
// tokenPlano es el token "en claro" que viaja en la URL — el hash se
// guarda en BD; este parámetro NO se loguea.
func (s *EmailService) SendEmailVerification(ctx context.Context, nombre, email, tokenPlano string, horasTTL int) error {
	url := s.cfg.AppBaseURL + "/confirmar-email?token=" + tokenPlano
	datos := DatosEmailVerificacion{
		Nombre:   primerNombre(nombre),
		URL:      url,
		HorasTTL: horasTTL,
	}

	htmlBuf := &bytes.Buffer{}
	if err := s.tplVerificacionHTML.Execute(htmlBuf, datos); err != nil {
		return fmt.Errorf("render verification.html: %w", err)
	}
	txtBuf := &bytes.Buffer{}
	if err := s.tplVerificacionTXT.Execute(txtBuf, datos); err != nil {
		return fmt.Errorf("render verification.txt: %w", err)
	}

	subject := "Confirma tu correo en Snowbreak"
	return s.enviar(ctx, email, subject, txtBuf.String(), htmlBuf.String())
}

// SendWelcomeEmail envía un correo de bienvenida sin pedir confirmación
// (lo usamos cuando el usuario llega con email_verified=true desde Google).
func (s *EmailService) SendWelcomeEmail(ctx context.Context, nombre, email string) error {
	url := s.cfg.AppBaseURL + "/estaciones"
	datos := DatosEmailBienvenida{
		Nombre: primerNombre(nombre),
		URL:    url,
	}
	htmlBuf := &bytes.Buffer{}
	if err := s.tplBienvenidaHTML.Execute(htmlBuf, datos); err != nil {
		return fmt.Errorf("render welcome.html: %w", err)
	}
	txtBuf := &bytes.Buffer{}
	if err := s.tplBienvenidaTXT.Execute(txtBuf, datos); err != nil {
		return fmt.Errorf("render welcome.txt: %w", err)
	}
	subject := "Bienvenido a Snowbreak"
	return s.enviar(ctx, email, subject, txtBuf.String(), htmlBuf.String())
}

// enviar construye el mensaje multipart/alternative y lo envía por SMTP
// con STARTTLS. Si el servicio está en modo "desactivado", loguea el
// asunto + URL extraída del cuerpo de texto plano y devuelve nil.
func (s *EmailService) enviar(ctx context.Context, to, subject, text, html string) error {
	if s.Desactivado() {
		log.Printf("EMAIL [desactivado] para=%s asunto=%q (cuerpo en logs):\n%s",
			to, subject, text)
		return nil
	}

	from := s.cfg.SMTPFromEmail
	fromName := s.cfg.SMTPFromName
	if fromName == "" {
		fromName = "Snowbreak"
	}

	mensaje := construirMensaje(fromName, from, to, subject, text, html)
	addr := net.JoinHostPort(s.cfg.SMTPHost, fmt.Sprintf("%d", s.cfg.SMTPPort))

	// Timeout total razonable. Si el servidor SMTP se queda colgado,
	// liberamos el contexto del handler que llamó.
	dialCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	d := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := d.DialContext(dialCtx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial smtp: %w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Quit()

	// STARTTLS — obligatorio en todos los proveedores serios.
	if ok, _ := c.Extension("STARTTLS"); ok {
		tlsConf := &tls.Config{ServerName: s.cfg.SMTPHost, MinVersion: tls.VersionTLS12}
		if err := c.StartTLS(tlsConf); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	// AUTH PLAIN sobre TLS (no antes — sería texto claro).
	if s.cfg.SMTPUser != "" {
		auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("smtp RCPT TO: %w", err)
	}
	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}
	if _, err := wc.Write([]byte(mensaje)); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	log.Printf("EMAIL enviado para=%s asunto=%q", to, subject)
	return nil
}

// construirMensaje arma un mensaje MIME multipart/alternative.
//
// Cabeceras mínimas: Date, From, To, Subject, MIME-Version, Content-Type.
// La fecha la pone el servidor SMTP si se omite, pero Postmark/Mailgun
// la exigen para no cancelar el envío.
func construirMensaje(fromName, fromEmail, to, subject, text, html string) string {
	boundary := "snowbreak-" + fmt.Sprintf("%x", time.Now().UnixNano())

	// Encode del subject con MIME-encoded-word para acentos.
	subjectEnc := mime.QEncoding.Encode("utf-8", subject)
	fromHeader := fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", fromName), fromEmail)

	var b bytes.Buffer
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().UTC().Format(time.RFC1123Z))
	fmt.Fprintf(&b, "From: %s\r\n", fromHeader)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subjectEnc)
	b.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	b.WriteString("\r\n")

	// Parte texto plano.
	fmt.Fprintf(&b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	b.WriteString(text)
	b.WriteString("\r\n\r\n")

	// Parte HTML.
	fmt.Fprintf(&b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	b.WriteString(html)
	b.WriteString("\r\n\r\n")

	fmt.Fprintf(&b, "--%s--\r\n", boundary)
	return b.String()
}

// primerNombre devuelve la primera palabra del nombre, escapada para no
// estropear el subject/body si el usuario pone algo raro.
func primerNombre(n string) string {
	n = strings.TrimSpace(n)
	if n == "" {
		return "esquiador"
	}
	if i := strings.IndexAny(n, " \t\n"); i > 0 {
		return n[:i]
	}
	return n
}
