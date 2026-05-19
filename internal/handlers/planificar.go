// Package handlers — pagina del asistente "Planifica tu estancia".
//
// Ruta:
//
//	GET /planificar-estancia → wizard de 4 pasos:
//	  1) elegir estacion (destino)
//	  2) elegir fechas y huespedes
//	  3) elegir alojamiento (de la estacion seleccionada)
//	  4) recomendacion de forfait (opcional, se anade al carrito existente)
//
// Toda la logica de pasos vive en el frontend (planificar.js): el
// servidor solo entrega la plantilla y los endpoints JSON:
//   - GET /api/estaciones                       → lista de estaciones
//   - GET /api/alojamientos/estacion/{id}       → lista de alojamientos
//   - POST /api/alojamientos/reservar           → confirma reserva
//
// La logica de carrito de forfaits no se toca: el asistente solo
// invoca el cart.js existente (que usa localStorage) para anadir el
// forfait al carrito y redirigir a /cesta.
package handlers

import (
	"net/http"

	"skihub/internal/models"
)

type datosPlanificar struct {
	Titulo      string
	Descripcion string
	Activa      string
	Usuario     *models.Usuario
}

// PlanificarEstancia responde a GET /planificar-estancia.
func (a *App) PlanificarEstancia(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/planificar-estancia" {
		a.NotFound(w, r)
		return
	}
	u := a.UsuarioActual(r)
	render(w, r, a.Plantillas, "planificar", datosPlanificar{
		Titulo:      "Planifica tu estancia — SnowBreak",
		Descripcion: "Crea tu escapada a la nieve paso a paso: elige estación, fechas, alojamiento y forfait en un solo flujo.",
		Activa:      "planificar",
		Usuario:     u,
	})
}
