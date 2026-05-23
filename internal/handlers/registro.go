package handlers

import (
	"errors"
	"log"
	"net/http"

	"skihub/internal/models"
	"skihub/internal/services"
)

type datosRegistro struct {
	Titulo      string
	Descripcion string
	Activa      string
	Error       string
	Form        formRegistro
	Usuario     *models.Usuario
	CSRF        string
}

type formRegistro struct {
	Nombre string
	Email  string
}

type datosRegistroOK struct {
	Titulo      string
	Descripcion string
	Activa      string
	Nombre      string
	Email       string
	// PendienteVerificar=true cuando el email aún no se ha confirmado
	// (registro local). En el caso de Google con email_verified=true no
	// se renderiza esta plantilla; se redirige al usuario logueado.
	PendienteVerificar bool
	Usuario            *models.Usuario
}

func (a *App) Registro(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.mostrarFormulario(w, r, datosRegistro{
			Titulo:      "Crear cuenta - SnowBreak",
			Descripcion: "Regístrate en SnowBreak.",
			Activa:      "registro",
			Usuario:     a.UsuarioActual(r),
			CSRF:        TokenCSRFActual(r),
		})
	case http.MethodPost:
		a.procesarRegistro(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
	}
}

func (a *App) mostrarFormulario(w http.ResponseWriter, r *http.Request, d datosRegistro) {
	render(w, r, a.Plantillas, "registro", d)
}

func (a *App) procesarRegistro(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ERROR parse form: %v", err)
		http.Error(w, "datos de formulario incorrectos", http.StatusBadRequest)
		return
	}

	datos := services.DatosRegistro{
		Nombre:    r.FormValue("nombre"),
		Email:     r.FormValue("email"),
		Password:  r.FormValue("password"),
		Password2: r.FormValue("password2"),
	}

	u, tokenPlano, err := a.UsuarioSvc.RegistrarConVerificacion(r.Context(), datos)
	if err != nil {
		if esErrorDeValidacion(err) {
			a.mostrarFormulario(w, r, datosRegistro{
				Titulo:      "Crear cuenta - SnowBreak",
				Descripcion: "Regístrate en SnowBreak.",
				Activa:      "registro",
				Error:       err.Error(),
				Form:        formRegistro{Nombre: datos.Nombre, Email: datos.Email},
				Usuario:     a.UsuarioActual(r),
				CSRF:        TokenCSRFActual(r),
			})
			return
		}
		log.Printf("ERROR registrando usuario: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	// Envío del email de verificación (24h TTL). Si falla NO rompemos el
	// flujo — el usuario ya se ha creado y podrá reenviar el correo más
	// tarde desde su cuenta.
	if a.EmailSvc != nil {
		if err := a.EmailSvc.SendEmailVerification(r.Context(), u.Nombre, u.Email, tokenPlano, 24); err != nil {
			log.Printf("EMAIL verificación fallo email=%s usuario=%d: %v",
				maskEmail(u.Email), u.ID, err)
			// Se sigue, pero indicamos en logs que el correo no se envió.
		}
	}

	render(w, r, a.Plantillas, "registro_ok", datosRegistroOK{
		Titulo:             "Registro completado - SnowBreak",
		Descripcion:        "Tu cuenta se ha creado correctamente en SnowBreak.",
		Activa:             "registro",
		Nombre:             u.Nombre,
		Email:              u.Email,
		PendienteVerificar: true,
		Usuario:            nil,
	})
}

func esErrorDeValidacion(err error) bool {
	return errors.Is(err, services.ErrNombreInvalido) ||
		errors.Is(err, services.ErrEmailInvalido) ||
		errors.Is(err, services.ErrPasswordDebil) ||
		errors.Is(err, services.ErrPasswordsNoCoinc) ||
		errors.Is(err, services.ErrEmailYaExiste)
}
