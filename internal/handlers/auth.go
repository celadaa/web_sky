package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"skihub/internal/models"
	"skihub/internal/services"
)

type datosLogin struct {
	Titulo      string
	Descripcion string
	Activa      string
	Error       string
	Mensaje     string
	Email       string
	Usuario     *models.Usuario
	CSRF        string
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if u := a.UsuarioActual(r); u != nil {
			http.Redirect(w, r, "/favoritos", http.StatusSeeOther)
			return
		}
		render(w, r, a.Plantillas, "login", datosLogin{
			Titulo:      "Iniciar sesión - SnowBreak",
			Descripcion: "Accede a tu cuenta de SnowBreak.",
			Activa:      "login",
			Mensaje:     limpiarMensaje(r.URL.Query().Get("mensaje")),
			CSRF:        TokenCSRFActual(r),
		})
	case http.MethodPost:
		a.procesarLogin(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
	}
}

func (a *App) procesarLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "datos de formulario incorrectos", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")

	u, err := a.UsuarioSvc.IniciarSesion(r.Context(), email, password)
	if err != nil {
		ip := ""
		if a.Cfg != nil {
			ip = IPCliente(r, a.Cfg.TrustProxy)
		}
		if errors.Is(err, services.ErrCredenciales) {
			log.Printf("AUTH: login fallido email=%q ip=%s", maskEmail(email), ip)
			render(w, r, a.Plantillas, "login", datosLogin{
				Titulo:      "Iniciar sesión - SnowBreak",
				Descripcion: "Accede a tu cuenta de SnowBreak.",
				Activa:      "login",
				Error:       "Correo o contraseña incorrectos.",
				Email:       email,
				CSRF:        TokenCSRFActual(r),
			})
			return
		}
		log.Printf("ERROR login interno email=%q ip=%s: %v", maskEmail(email), ip, err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	sesion, err := a.SesionSvc.Crear(r.Context(), u.ID)
	if err != nil {
		log.Printf("ERROR creando sesión usuario=%d: %v", u.ID, err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieSesion,
		Value:    sesion.Token,
		Path:     "/",
		Expires:  sesion.ExpiraEn,
		HttpOnly: true,
		Secure:   a.CookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	if a.Cfg != nil {
		log.Printf("AUTH: login OK usuario=%d ip=%s", u.ID, IPCliente(r, a.Cfg.TrustProxy))
	}

	// Solo redirigimos a destinos internos (anti Open Redirect).
	destino := r.FormValue("redirigir")
	if !EsRedirectInterno(destino) {
		destino = "/estaciones"
	}
	if destino == "" {
		destino = "/estaciones"
	}
	http.Redirect(w, r, destino, http.StatusSeeOther)
}

type datosCambiarPwd struct {
	Titulo      string
	Descripcion string
	Activa      string
	Error       string
	Mensaje     string
	Usuario     *models.Usuario
	CSRF        string
}

func (a *App) CambiarPassword(w http.ResponseWriter, r *http.Request) {
	u := a.UsuarioActual(r)
	if u == nil {
		http.Redirect(w, r, "/login?mensaje=Inicia+sesi%C3%B3n+para+cambiar+tu+contrase%C3%B1a", http.StatusSeeOther)
		return
	}
	switch r.Method {
	case http.MethodGet:
		render(w, r, a.Plantillas, "cambiar_password", datosCambiarPwd{
			Titulo:      "Cambiar contraseña - SnowBreak",
			Descripcion: "Actualiza la contraseña de tu cuenta.",
			Activa:      "cambiar_password",
			Mensaje:     limpiarMensaje(r.URL.Query().Get("mensaje")),
			Usuario:     u,
			CSRF:        TokenCSRFActual(r),
		})
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "datos de formulario incorrectos", http.StatusBadRequest)
			return
		}
		actual := r.FormValue("password_actual")
		nueva := r.FormValue("password_nueva")
		nueva2 := r.FormValue("password_nueva2")

		err := a.UsuarioSvc.CambiarPassword(r.Context(), u.ID, actual, nueva, nueva2)
		if err != nil {
			msg := "Error al cambiar la contraseña."
			switch {
			case errors.Is(err, services.ErrPasswordActual):
				msg = "La contraseña actual no es correcta."
			case errors.Is(err, services.ErrPasswordDebil):
				msg = "La contraseña nueva debe tener al menos 8 caracteres."
			case errors.Is(err, services.ErrPasswordsNoCoinc):
				msg = "La contraseña nueva y su confirmación no coinciden."
			case errors.Is(err, services.ErrPasswordIgual):
				msg = "La contraseña nueva debe ser distinta de la actual."
			default:
				log.Printf("ERROR cambiando password usuario id=%d: %v", u.ID, err)
			}
			render(w, r, a.Plantillas, "cambiar_password", datosCambiarPwd{
				Titulo:      "Cambiar contraseña - SnowBreak",
				Descripcion: "Actualiza la contraseña de tu cuenta.",
				Activa:      "cambiar_password",
				Error:       msg,
				Usuario:     u,
				CSRF:        TokenCSRFActual(r),
			})
			return
		}
		// Tras un cambio de contraseña exitoso conviene invalidar las sesiones
		// existentes para forzar un re-login en otros dispositivos. Como la
		// implementación actual no tiene ese hook, al menos rotamos la sesión
		// activa eliminando la actual y creando una nueva.
		if c, err := r.Cookie(CookieSesion); err == nil && a.SesionSvc != nil {
			_ = a.SesionSvc.Cerrar(r.Context(), c.Value)
		}
		if a.SesionSvc != nil {
			if nueva, err2 := a.SesionSvc.Crear(r.Context(), u.ID); err2 == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     CookieSesion,
					Value:    nueva.Token,
					Path:     "/",
					Expires:  nueva.ExpiraEn,
					HttpOnly: true,
					Secure:   a.CookieSecure(),
					SameSite: http.SameSiteLaxMode,
				})
			}
		}
		http.Redirect(w, r, "/cambiar-password?mensaje=Contrase%C3%B1a+actualizada+correctamente", http.StatusSeeOther)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
	}
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if c, err := r.Cookie(CookieSesion); err == nil {
		_ = a.SesionSvc.Cerrar(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieSesion,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   a.CookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/login?mensaje=Sesi%C3%B3n+cerrada+correctamente", http.StatusSeeOther)
}

// limpiarMensaje recorta y trunca un texto para evitar que un atacante
// cuelgue payloads largos en la query string. html/template ya escapa
// el contenido, pero limitar la longitud reduce ruido en la UI y logs.
func limpiarMensaje(s string) string {
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}

// maskEmail oculta parte del local-part para no escribir emails en
// claro en los logs (privacidad / cumplimiento RGPD).
func maskEmail(e string) string {
	at := -1
	for i, c := range e {
		if c == '@' {
			at = i
			break
		}
	}
	if at <= 1 {
		return "***"
	}
	if at > 60 {
		return "***"
	}
	local := e[:at]
	dom := e[at:]
	if len(local) <= 2 {
		return "*" + dom
	}
	return local[:1] + "***" + local[len(local)-1:] + dom
}
