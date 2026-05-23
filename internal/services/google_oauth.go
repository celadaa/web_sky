// Package services — login con Google vía OAuth 2.0 + OpenID Connect.
//
// Diseño:
//   - El intercambio code → tokens se hace en el SERVIDOR. El navegador
//     nunca ve client_secret ni access_token.
//   - El ID token se verifica con go-oidc (firma RS256 contra JWKS de
//     Google, issuer accounts.google.com, audience == GOOGLE_CLIENT_ID,
//     expiración). No nos fiamos de los campos que vienen en el code
//     redirect.
//   - Scopes mínimos: "openid email profile". NO pedimos acceso a Gmail,
//     contactos, ni nada sensible.
//   - access_token y refresh_token NO se persisten (no los necesitamos
//     para nada después del primer callback).
//
// Política de vinculación:
//
//   - Si hay un usuario con ese google_id → login directo.
//   - Si no, pero hay un usuario con ese email Y email_verified=true del
//     ID token Y email_verified=true en BD (o lo marcamos al vincular):
//     se vincula el google_id a esa cuenta.
//   - Si el email del ID token NO está verificado, NO vinculamos a una
//     cuenta existente (anti account-takeover) — registramos un error
//     y mostramos al usuario que verifique primero.
//   - Si no existe ningún usuario con ese email: se crea uno nuevo con
//     auth_provider='google'.
package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"

	"skihub/internal/config"
	"skihub/internal/models"
	"skihub/internal/repository"
)

// GoogleIssuer es el issuer estándar de Google OIDC. Se valida en cada
// callback contra el campo `iss` del ID token.
const GoogleIssuer = "https://accounts.google.com"

// Errores expuestos al handler. El handler los traduce a mensajes de
// usuario sin filtrar detalles internos.
var (
	ErrGoogleDesactivado     = errors.New("login con Google no está configurado")
	ErrGoogleEstadoInvalido  = errors.New("estado OAuth inválido (posible CSRF)")
	ErrGoogleCodigoInvalido  = errors.New("código de autorización inválido")
	ErrGoogleIDTokenAusente  = errors.New("Google no devolvió id_token")
	ErrGoogleIDTokenInvalido = errors.New("id_token inválido")
	ErrGoogleEmailNoVerif    = errors.New("Google indica que el email no está verificado")
	ErrGoogleAccountTakeover = errors.New("ya existe una cuenta con ese email; verifica primero el email en Snowbreak antes de vincular Google")
)

// GoogleOAuthService implementa el flujo OIDC con Google.
type GoogleOAuthService struct {
	cfg      *config.Config
	oauth    *oauth2.Config
	verifier *oidc.IDTokenVerifier
	repo     *repository.UsuarioRepo
}

// NuevoGoogleOAuthService inicializa el cliente OAuth y el verificador OIDC.
//
// Devuelve (nil, nil) si Google no está configurado (ClientID o Secret
// vacíos). El llamador (main.go) debe comprobar el nil y desactivar las
// rutas correspondientes.
func NuevoGoogleOAuthService(ctx context.Context, cfg *config.Config, repo *repository.UsuarioRepo) (*GoogleOAuthService, error) {
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		return nil, nil //nolint:nilnil // contrato explícito: nil = "no configurado"
	}
	provider, err := oidc.NewProvider(ctx, GoogleIssuer)
	if err != nil {
		return nil, fmt.Errorf("oidc provider: %w", err)
	}
	conf := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Endpoint:     googleoauth.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.GoogleClientID})
	return &GoogleOAuthService{
		cfg:      cfg,
		oauth:    conf,
		verifier: verifier,
		repo:     repo,
	}, nil
}

// AuthCodeURL devuelve la URL a la que redirigir al usuario para que
// inicie sesión en accounts.google.com. `state` debe ser un valor
// aleatorio que el handler también guardará en una cookie para validar
// el callback (anti CSRF).
func (s *GoogleOAuthService) AuthCodeURL(state string) string {
	return s.oauth.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// GenerarEstado produce 32 bytes aleatorios codificados en base64
// URL-safe para usar como parámetro `state` del OAuth flow.
func GenerarEstado() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ClaimsGoogle son los campos que extraemos del ID token verificado.
// Aceptamos exactamente los del scope "openid email profile".
type ClaimsGoogle struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// IntercambiarYVerificar canjea el `code` por tokens y valida el ID token.
//
// La firma se valida contra JWKS de Google (cacheado por go-oidc),
// issuer == GoogleIssuer y audience == clientID. Si todo OK extrae los
// claims que nos interesan.
func (s *GoogleOAuthService) IntercambiarYVerificar(ctx context.Context, code string) (*ClaimsGoogle, error) {
	if code == "" {
		return nil, ErrGoogleCodigoInvalido
	}
	tok, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange: %w", err)
	}
	raw, ok := tok.Extra("id_token").(string)
	if !ok || raw == "" {
		return nil, ErrGoogleIDTokenAusente
	}
	idTok, err := s.verifier.Verify(ctx, raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoogleIDTokenInvalido, err)
	}
	var c ClaimsGoogle
	if err := idTok.Claims(&c); err != nil {
		return nil, fmt.Errorf("claims: %w", err)
	}
	c.Email = strings.ToLower(strings.TrimSpace(c.Email))
	if c.Sub == "" || c.Email == "" {
		return nil, ErrGoogleIDTokenInvalido
	}
	return &c, nil
}

// ResolverUsuarioDesdeGoogle aplica la política de vinculación descrita
// arriba y devuelve el usuario con el que se debe iniciar sesión, junto
// con un booleano `creado` para que el handler decida si enviar email
// de bienvenida.
//
// Casos:
//
//   - usuario con google_id == claims.Sub: login directo.
//   - usuario con email == claims.Email Y claims.EmailVerified == true:
//     se vincula google_id y se marca email_verified=true; login.
//   - usuario con email == claims.Email Y claims.EmailVerified == false:
//     ErrGoogleAccountTakeover. No vinculamos para evitar que alguien con
//     un email no verificado en Google nos secuestre la cuenta.
//   - sin usuario con ese email: se crea uno nuevo con auth_provider='google'.
func (s *GoogleOAuthService) ResolverUsuarioDesdeGoogle(ctx context.Context, c *ClaimsGoogle) (u *models.Usuario, creado bool, err error) {
	if c == nil {
		return nil, false, errors.New("claims nil")
	}
	// 1) ¿Ya hay alguien vinculado a este google_id?
	u, err = s.repo.BuscarPorGoogleID(ctx, c.Sub)
	if err == nil {
		// Refrescamos avatar si Google lo trajo y nosotros no tenemos uno.
		if u.AvatarURL == nil && c.Picture != "" {
			pic := c.Picture
			// Mejor esfuerzo: si falla, el login no se interrumpe.
			_ = s.repo.VincularGoogle(ctx, u.ID, c.Sub, &pic)
		}
		return u, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, fmt.Errorf("buscar por google_id: %w", err)
	}

	// 2) ¿Hay un usuario con ese email?
	existente, err := s.repo.BuscarPorEmailCompleto(ctx, c.Email)
	switch {
	case err == nil:
		// Anti account-takeover: sólo vinculamos si Google confirma que
		// el email está verificado de su lado.
		if !c.EmailVerified {
			return nil, false, ErrGoogleAccountTakeover
		}
		pic := nilSiVacio(c.Picture)
		if err := s.repo.VincularGoogle(ctx, existente.ID, c.Sub, pic); err != nil {
			return nil, false, fmt.Errorf("vincular google: %w", err)
		}
		// Marcamos verified si todavía no lo estaba.
		if !existente.EmailVerificado {
			if err := s.repo.MarcarEmailVerificado(ctx, existente.ID); err != nil {
				return nil, false, fmt.Errorf("marcar verified: %w", err)
			}
			existente.EmailVerificado = true
		}
		existente.GoogleID = &c.Sub
		if pic != nil {
			existente.AvatarURL = pic
		}
		return existente, false, nil
	case errors.Is(err, sql.ErrNoRows):
		// 3) No existe → creamos cuenta nueva.
		pic := nilSiVacio(c.Picture)
		sub := c.Sub
		nuevo := &models.Usuario{
			Nombre:          nombreSeguro(c.Name, c.Email),
			Email:           c.Email,
			EmailVerificado: c.EmailVerified,
			GoogleID:        &sub,
			AvatarURL:       pic,
			AuthProvider:    "google",
		}
		if err := s.repo.CrearGoogle(ctx, nuevo); err != nil {
			return nil, false, fmt.Errorf("crear usuario google: %w", err)
		}
		return nuevo, true, nil
	default:
		return nil, false, fmt.Errorf("buscar por email: %w", err)
	}
}

// nombreSeguro devuelve un nombre presentable que cumpla la restricción
// CHECK users_name_len (entre 2 y 60 caracteres). Si Google no manda
// `name`, caemos al local-part del email; si tampoco eso es válido,
// usamos un genérico para que la fila se inserte.
func nombreSeguro(name, email string) string {
	n := strings.TrimSpace(name)
	if l := lenRunes(n); l >= 2 && l <= 60 {
		return n
	}
	if at := strings.IndexByte(email, '@'); at > 1 {
		local := email[:at]
		if l := lenRunes(local); l >= 2 && l <= 60 {
			return local
		}
	}
	return "Esquiador"
}

func lenRunes(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func nilSiVacio(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
