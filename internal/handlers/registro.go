package handlers

import (
	"errors"
	"log"
	"net/http"

	"skihub/internal/models"
	"skihub/internal/services"
)

// datosRegistro son los datos que necesita la plantilla registro.tmpl.
// Incluimos Form para poder "re-mostrar" lo que el usuario había escrito
// cuando la validación falla.
type datosRegistro struct {
	Titulo      string
	Descripcion string
	Activa      string
	Error       string
	Form        formRegistro
	Usuario     *models.Usuario
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
	Usuario     *models.Usuario
}

// Registro es el manejador combinado para GET y POST de /registro.
// - GET  → muestra el formulario vacío.
// - POST → valida, registra y muestra la página de éxito (o vuelve al formulario con error).
func (a *App) Registro(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.mostrarFormulario(w, r, datosRegistro{
			Titulo:      "Crear cuenta - SkiHub",
			Descripcion: "Regístrate en SkiHub.",
			Activa:      "registro",
			Usuario:     a.UsuarioActual(r),
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
	// Analizar los campos del formulario (Content-Type: application/x-www-form-urlencoded)
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

	log.Printf("POST /registro email=%s", datos.Email)

	u, err := a.UsuarioSvc.Registrar(datos)
	if err != nil {
		// Errores "de negocio": se vuelve al formulario con el mensaje.
		if esErrorDeValidacion(err) {
			a.mostrarFormulario(w, r, datosRegistro{
				Titulo:      "Crear cuenta - SkiHub",
				Descripcion: "Regístrate en SkiHub.",
				Activa:      "registro",
				Error:       err.Error(),
				Form:        formRegistro{Nombre: datos.Nombre, Email: datos.Email},
				Usuario:     a.UsuarioActual(r),
			})
			return
		}
		// Error inesperado: se registra y se devuelve 500.
		log.Printf("ERROR registrando usuario: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	log.Printf("Usuario creado id=%d email=%s", u.ID, u.Email)

	// Página de confirmación. Desde aquí invitamos a iniciar sesión.
	render(w, r, a.Plantillas, "registro_ok", datosRegistroOK{
		Titulo:      "Registro completado - SkiHub",
		Descripcion: "Tu cuenta se ha creado correctamente en SkiHub.",
		Activa:      "registro",
		Nombre:      u.Nombre,
		Email:       u.Email,
		Usuario:     nil, // recién registrado aún no ha iniciado sesión
	})
}

func esErrorDeValidacion(err error) bool {
	return errors.Is(err, services.ErrNombreInvalido) ||
		errors.Is(err, services.ErrEmailInvalido) ||
		errors.Is(err, services.ErrPasswordDebil) ||
		errors.Is(err, services.ErrPasswordsNoCoinc) ||
		errors.Is(err, services.ErrEmailYaExiste)
}
