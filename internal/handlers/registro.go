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

	u, err := a.UsuarioSvc.Registrar(datos)
	if err != nil {
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
		log.Printf("ERROR registrando usuario: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	render(w, r, a.Plantillas, "registro_ok", datosRegistroOK{
		Titulo:      "Registro completado - SkiHub",
		Descripcion: "Tu cuenta se ha creado correctamente en SkiHub.",
		Activa:      "registro",
		Nombre:      u.Nombre,
		Email:       u.Email,
		Usuario:     nil,
	})
}

func esErrorDeValidacion(err error) bool {
	return errors.Is(err, services.ErrNombreInvalido) ||
		errors.Is(err, services.ErrEmailInvalido) ||
		errors.Is(err, services.ErrPasswordDebil) ||
		errors.Is(err, services.ErrPasswordsNoCoinc) ||
		errors.Is(err, services.ErrEmailYaExiste)
}
