package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"skihub/internal/models"
)

type datosEstaciones struct {
	Titulo      string
	Descripcion string
	Activa      string
	Estaciones  []models.Estacion
	Usuario     *models.Usuario
}

type datosEstacion struct {
	Titulo      string
	Descripcion string
	Activa      string
	Estacion    *models.Estacion
	Usuario     *models.Usuario
}

// Estaciones responde a GET /estaciones con el listado completo.
// Si hay usuario autenticado, marca las que son favoritas suyas.
func (a *App) Estaciones(w http.ResponseWriter, r *http.Request) {
	u := a.UsuarioActual(r)
	var uid int64
	if u != nil {
		uid = u.ID
	}
	lista, err := a.EstacionSvc.Listar(uid)
	if err != nil {
		log.Printf("ERROR listar estaciones: %v", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	render(w, r, a.Plantillas, "estaciones", datosEstaciones{
		Titulo:      "Estaciones - SkiHub",
		Descripcion: "Listado de estaciones de esquí disponibles en SkiHub.",
		Activa:      "estaciones",
		Estaciones:  lista,
		Usuario:     u,
	})
}

// Estacion responde a GET /estacion/{id} con la ficha de una estación concreta.
// Usamos el enrutador de la biblioteca estándar con prefijo "/estacion/".
func (a *App) Estacion(w http.ResponseWriter, r *http.Request) {
	// Extraer el ID del path: /estacion/3 -> "3"
	idStr := strings.TrimPrefix(r.URL.Path, "/estacion/")
	if idStr == "" || strings.Contains(idStr, "/") {
		a.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		a.NotFound(w, r)
		return
	}
	u := a.UsuarioActual(r)
	var uid int64
	if u != nil {
		uid = u.ID
	}
	e, err := a.EstacionSvc.Obtener(id, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.NotFound(w, r)
			return
		}
		log.Printf("ERROR obtener estación %d: %v", id, err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	render(w, r, a.Plantillas, "estacion", datosEstacion{
		Titulo:      e.Nombre + " - SkiHub",
		Descripcion: "Estado y condiciones de la estación de esquí de " + e.Nombre + ".",
		Activa:      "estaciones",
		Estacion:    e,
		Usuario:     u,
	})
}
