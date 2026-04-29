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
		Titulo:      "Página no encontrada - SkiHub",
		Descripcion: "La página solicitada no existe.",
		Activa:      "",
		Codigo:      http.StatusNotFound,
		Mensaje:     "No encontrado",
		Detalle:     "La página que buscas no existe o se ha movido.",
		Usuario:     a.UsuarioActual(r),
	})
}
