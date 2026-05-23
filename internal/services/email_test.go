package services

import (
	"strings"
	"testing"
)

// TestPrimerNombreFallback comprueba el fallback "esquiador" para nombres
// vacíos. Sin esto, los emails dirían "Hola ," que queda raro.
func TestPrimerNombreFallback(t *testing.T) {
	cases := map[string]string{
		"":                    "esquiador",
		" ":                   "esquiador",
		"María":               "María",
		"María José":          "María",
		"\tCelada\nOrtega":    "Celada",
		"Juan Carlos Pérez":   "Juan",
		" Juan ":              "Juan",
	}
	for in, want := range cases {
		got := primerNombre(in)
		if got != want {
			t.Errorf("primerNombre(%q) = %q, quería %q", in, got, want)
		}
	}
}

// TestConstruirMensajeIncluyeAmbasPartes verifica que el cuerpo MIME
// contiene la sección text/plain y text/html separadas por el boundary.
func TestConstruirMensajeIncluyeAmbasPartes(t *testing.T) {
	msg := construirMensaje(
		"Snowbreak",
		"no-reply@snowbreak.es",
		"usuario@ejemplo.com",
		"Asunto con eñe",
		"cuerpo plano\n",
		"<p>cuerpo html</p>",
	)
	if !strings.Contains(msg, "multipart/alternative") {
		t.Error("mensaje no declara multipart/alternative")
	}
	if !strings.Contains(msg, "Content-Type: text/plain") {
		t.Error("mensaje no incluye parte text/plain")
	}
	if !strings.Contains(msg, "Content-Type: text/html") {
		t.Error("mensaje no incluye parte text/html")
	}
	if !strings.Contains(msg, "cuerpo plano") {
		t.Error("falta el cuerpo plano en el mensaje")
	}
	if !strings.Contains(msg, "<p>cuerpo html</p>") {
		t.Error("falta el cuerpo html en el mensaje")
	}
	// El subject con tilde debe ir QP-encoded.
	if !strings.Contains(msg, "Subject: =?utf-8?q?") {
		t.Errorf("subject no codificado en MIME QEncoding: %q", subjectFromMessage(msg))
	}
}

func subjectFromMessage(msg string) string {
	const k = "Subject: "
	i := strings.Index(msg, k)
	if i < 0 {
		return ""
	}
	rest := msg[i+len(k):]
	if j := strings.Index(rest, "\r\n"); j >= 0 {
		return rest[:j]
	}
	return rest
}
