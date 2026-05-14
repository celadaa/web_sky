// Package handlers — endpoint de "refresco" del parte de nieve.
//
// GET /api/estacion/{id}/parte
//
// Devuelve los datos completos de una estación en JSON, recalculando los
// campos derivados (nieve mín/máx, viento, fecha de última actualización)
// con la lógica determinista del servicio. Sirve para que el botón
// "Actualizar parte" de la ficha de estación pueda refrescar los valores
// sin recargar la página entera.
package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"database/sql"
)

// parteJSON es el DTO que consume parte-refresh.js
type parteJSON struct {
	ID                int64  `json:"id"`
	Estado            string `json:"estado"`
	Temperatura       int    `json:"temperatura"`
	NieveMin          int    `json:"nieve_min"`
	NieveMax          int    `json:"nieve_max"`
	NieveBase         int    `json:"nieve_base"`
	NieveNueva        int    `json:"nieve_nueva"`
	Viento            string `json:"viento"`
	UltimaNevada      string `json:"ultima_nevada"`
	PistasAbiertas    int    `json:"pistas_abiertas"`
	PistasTotales     int    `json:"pistas_totales"`
	RemontesOp        int    `json:"remontes_op"`
	RemontesTot       int    `json:"remontes_tot"`
	ParteHora         string `json:"parte_hora"`
	ParteActualizado  string `json:"parte_actualizado"`
	ParteHaceMinutos  int    `json:"parte_hace_minutos"`
}

// ApiParteEstacion responde a GET /api/estacion/{id}/parte
func (a *App) ApiParteEstacion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		escribirError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	// La ruta es "/api/estacion/{id}/parte"
	path := strings.TrimPrefix(r.URL.Path, "/api/estacion/")
	if !strings.HasSuffix(path, "/parte") {
		escribirError(w, http.StatusNotFound, "endpoint no encontrado")
		return
	}
	idStr := strings.TrimSuffix(path, "/parte")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		escribirError(w, http.StatusBadRequest, "id inválido")
		return
	}

	e, err := a.EstacionSvc.Obtener(r.Context(), id, 0)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			escribirError(w, http.StatusNotFound, "estación no encontrada")
			return
		}
		log.Printf("ERROR API parte estación %d: %v", id, err)
		escribirError(w, http.StatusInternalServerError, "no hemos podido cargar el parte")
		return
	}

	out := parteJSON{
		ID:               e.ID,
		Estado:           e.EstadoTexto(),
		Temperatura:      e.Temperatura,
		NieveMin:         e.NieveMin,
		NieveMax:         e.NieveMax,
		NieveBase:        e.NieveBase,
		NieveNueva:       e.NieveNueva,
		Viento:           e.Viento,
		UltimaNevada:     e.UltimaNevada,
		PistasAbiertas:   e.PistasAbiertas,
		PistasTotales:    e.PistasTotales,
		RemontesOp:       e.RemontesOp,
		RemontesTot:      e.RemontesTot,
		ParteHora:        e.ParteHora(),
		ParteActualizado: e.ParteActualizadoTexto(),
		ParteHaceMinutos: e.ParteHaceMinutos(),
	}
	escribirJSON(w, http.StatusOK, out)
}
