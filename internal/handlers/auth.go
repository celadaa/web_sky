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
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if u := a.UsuarioActual(r); u != nil {
			http.Redirect(w, r, "/favoritos", http.StatusSeeOther)
			return
		}
		render(w, r, a.Plantillas, "login", datosLogin{
			Titulo:      "Iniciar sesión - Snowbreak",
			Descripcion: "Accede a tu cuenta de Snowbreak.",
			Activa:      "login",
			Mensaje:     r.URL.Query().Get("mensaje"),
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

	u, err := a.UsuarioSvc.IniciarSesion(email, password)
	if err != nil {
		if errors.Is(err, services.ErrCredenciales) {
			render(w, r, a.Plantillas, "login", datosLogin{
				Titulo:      "Iniciar sesión - Snowbreak",
				Descripcion: "Accede a tu cuenta de Snowbreak.",
				Activa:      "login",
				Error:       "Correo o contraseña incorrectos.",
				Email:       email,
			})
			return
		}
		log.Printf("ERROR login: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	sesion, err := a.SesionSvc.Crear(u.ID)
	if err != nil {
		log.Printf("ERROR creando sesión: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieSesion,
		Value:    sesion.Token,
		Path:     "/",
		Expires:  sesion.ExpiraEn,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/estaciones", http.StatusSeeOther)
}

type datosCambiarPwd struct {
	Titulo      string
	Descripcion string
	Activa      string
	Error       string
	Mensaje     string
	Usuario     *models.Usuario
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
			Titulo:      "Cambiar contraseña - Snowbreak",
			Descripcion: "Actualiza la contraseña de tu cuenta.",
			Activa:      "cambiar_password",
			Mensaje:     r.URL.Query().Get("mensaje"),
			Usuario:     u,
		})
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "datos de formulario incorrectos", http.StatusBadRequest)
			return
		}
		actual := r.FormValue("password_actual")
		nueva := r.FormValue("password_nueva")
		nueva2 := r.FormValue("password_nueva2")

		err := a.UsuarioSvc.CambiarPassword(u.ID, actual, nueva, nueva2)
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
				Titulo:      "Cambiar contraseña - Snowbreak",
				Descripcion: "Actualiza la contraseña de tu cuenta.",
				Activa:      "cambiar_password",
				Error:       msg,
				Usuario:     u,
			})
			return
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
		_ = a.SesionSvc.Cerrar(c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieSesion,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/login?mensaje=Sesi%C3%B3n+cerrada+correctamente", http.StatusSeeOther)
}
