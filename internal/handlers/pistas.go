package handlers

import (
	"net/http"

	"skihub/internal/models"
)

// datosPistas alimenta la plantilla pistas.tmpl. La carga real de
// estaciones se hace por JS desde /api/nieve/estaciones para que la
// petición ya pueda incluir las coordenadas del usuario tras pedir
// permiso de geolocalización.
type datosPistas struct {
	Titulo      string
	Descripcion string
	Activa      string
	Usuario     *models.Usuario
}

// Pistas → GET /pistas: página con estado en directo de las pistas.
func (a *App) Pistas(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/pistas" {
		a.NotFound(w, r)
		return
	}
	render(w, r, a.Plantillas, "pistas", datosPistas{
		Titulo:      "Estado de pistas en directo - SnowBreak",
		Descripcion: "Consulta el estado de apertura, kilómetros abiertos, remontes y nieve de las estaciones de esquí más cercanas a tu ubicación.",
		Activa:      "pistas",
		Usuario:     a.UsuarioActual(r),
	})
}
