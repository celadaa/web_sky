package handlers

import (
	"net/http"

	"skihub/internal/models"
)

type datosError struct {
	Titulo      string
	Descripcion string
	Activa      string
	Codigo      int
	Mensaje     string
	Detalle     string
	Usuario     *models.Usuario
}

// NotFound responde con una página 404 usando la plantilla error.tmpl.
func (a *App) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	render(w, r, a.Plantillas, "error", datosError{
		Titulo:      "404 · No hemos encontrado esta pista | SnowBreak",
		Descripcion: "La ruta que buscas no existe o se ha cubierto de nieve. Vuelve al inicio o explora las estaciones disponibles en SnowBreak.",
		Activa:      "",
		Codigo:      http.StatusNotFound,
		Mensaje:     "No hemos encontrado esta pista",
		Detalle:     "La ruta que buscas no existe o se ha cubierto de nieve.",
		Usuario:     a.UsuarioActual(r),
	})
}
