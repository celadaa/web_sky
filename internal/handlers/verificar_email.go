// Package handlers — confirmación de email vía token enviado por correo.
//
// Endpoint:
//
//	GET /confirmar-email?token=...
//
// Resultado:
//
//   - Token válido: marca el usuario como verificado y renderiza la pantalla
//     de éxito.
//   - Token inválido / caducado: renderiza la pantalla de error sin filtrar
//     cuál de las dos cosas ocurrió.
//
// No requiere sesión: el propio token es el "factor de autenticación" para
// esta operación. Por eso lo enviamos por un canal externo (email) y lo
// hacemos de un solo uso (al confirmar se borra de la BD).
package handlers

import (
	"errors"
	"log"
	"net/http"

	"skihub/internal/models"
	"skihub/internal/services"
)

type datosConfirmar struct {
	Titulo      string
	Descripcion string
	Activa      string
	Nombre      string
	Email       string
	Error       string
	Usuario     *models.Usuario
}

// ConfirmarEmail procesa el token de verificación.
func (a *App) ConfirmarEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		a.renderConfirmarError(w, r, "El enlace de confirmación no es válido.")
		return
	}

	u, err := a.UsuarioSvc.ConfirmarEmail(r.Context(), token)
	if err != nil {
		if errors.Is(err, services.ErrTokenInvalido) {
			a.renderConfirmarError(w, r, "El enlace ha caducado o ya se ha usado. Inicia sesión y pide reenvío si lo necesitas.")
			return
		}
		log.Printf("ERROR confirmando email: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	log.Printf("AUTH: email confirmado usuario=%d", u.ID)
	render(w, r, a.Plantillas, "confirmar_email_ok", datosConfirmar{
		Titulo:      "Email confirmado - SnowBreak",
		Descripcion: "Has confirmado tu correo en SnowBreak.",
		Activa:      "login",
		Nombre:      u.Nombre,
		Email:       u.Email,
		Usuario:     a.UsuarioActual(r),
	})
}

func (a *App) renderConfirmarError(w http.ResponseWriter, r *http.Request, mensaje string) {
	w.WriteHeader(http.StatusBadRequest)
	render(w, r, a.Plantillas, "confirmar_email_error", datosConfirmar{
		Titulo:      "Enlace no válido - SnowBreak",
		Descripcion: "El enlace de confirmación no es válido.",
		Activa:      "login",
		Error:       mensaje,
		Usuario:     a.UsuarioActual(r),
	})
}
