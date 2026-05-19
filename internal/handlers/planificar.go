package handlers

import (
	"log"
	"net/http"

	"skihub/internal/models"
)

// datosPlanificar agrupa los datos para la plantilla planificar.tmpl.
type datosPlanificar struct {
	Titulo      string
	Descripcion string
	Activa      string
	Usuario     *models.Usuario
	Estaciones  []models.Estacion
}

// PlanificarEstancia responde a GET /planificar-estancia con el wizard
// del configurador de viaje + ticket premium en tiempo real.
//
// El backend se mantiene mínimo: solo sirve la plantilla con el listado
// de estaciones (usamos las mismas que la home, así no añadimos APIs).
// La lógica del wizard, del ticket y de localStorage vive 100% en
// /static/js/planificar.js y /static/js/trip-planner-ticket.js.
func (a *App) PlanificarEstancia(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/planificar-estancia" {
		a.NotFound(w, r)
		return
	}
	ctx := r.Context()
	u := a.UsuarioActual(r)
	var uid int64
	if u != nil {
		uid = u.ID
	}
	estaciones, err := a.EstacionSvc.Listar(ctx, uid)
	if err != nil {
		log.Printf("ERROR planificar listar estaciones: %v", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	render(w, r, a.Plantillas, "planificar", datosPlanificar{
		Titulo:      "SnowBreak | Planifica tu viaje a la nieve",
		Descripcion: "Construye tu viaje paso a paso: estación, fechas, alojamiento y forfait. Ticket premium en tiempo real con todo lo que llevas.",
		Activa:      "planificar",
		Usuario:     u,
		Estaciones:  estaciones,
	})
}
