package handlers

import (
	"log"
	"net/http"

	"skihub/internal/models"
)

// datosHome es el struct que se pasa a la plantilla index.tmpl.
type datosHome struct {
	Titulo            string
	Descripcion       string
	Activa            string
	Estaciones        []models.Estacion
	MasCercana        *models.Estacion
	MasLejana         *models.Estacion
	DistanciaPromedio float64
	Usuario           *models.Usuario
}

// Home responde a GET / con el listado de estaciones más cercanas y el resumen.
func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		a.NotFound(w, r)
		return
	}
	cercana, lejana, promedio, _, err := a.EstacionSvc.ResumenHome()
	if err != nil {
		log.Printf("ERROR resumen home: %v", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	u := a.UsuarioActual(r)
	var uid int64
	if u != nil {
		uid = u.ID
	}
	estaciones, err := a.EstacionSvc.Listar(uid)
	if err != nil {
		log.Printf("ERROR listar estaciones: %v", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	render(w, r, a.Plantillas, "index", datosHome{
		Titulo:            "Snowbreak - Inicio | Encuentra tu estación de esquí",
		Descripcion:       "Encuentra tu estación de esquí perfecta, información meteorológica y distancias en Snowbreak.",
		Activa:            "inicio",
		Estaciones:        estaciones,
		MasCercana:        cercana,
		MasLejana:         lejana,
		DistanciaPromedio: promedio,
		Usuario:           u,
	})
}
