// Package handlers — alojamientos cercanos a estaciones y reservas.
//
// Rutas que añade este archivo:
//
//   GET  /alojamiento/{id}           → página de detalle + formulario reserva
//   POST /api/alojamientos/reservar  → confirmar reserva (requiere login)
//
// La sección "Alojamientos cercanos" dentro de la ficha de estación no
// requiere handler propio: el handler de Estacion ya carga la lista vía
// AlojamientoSvc y la pasa a la plantilla.
package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"skihub/internal/models"
	"skihub/internal/services"
)

// datosAlojamiento es el struct que recibe la plantilla alojamiento.tmpl.
type datosAlojamiento struct {
	Titulo       string
	Descripcion  string
	Activa       string
	Alojamiento  *models.Alojamiento
	Usuario      *models.Usuario
	YaLogueado   bool // azúcar para la plantilla
}

// Alojamiento responde a GET /alojamiento/{id} con la ficha del
// alojamiento y un formulario para reservar noches.
func (a *App) Alojamiento(w http.ResponseWriter, r *http.Request) {
	if a.AlojamientoSvc == nil {
		// Si el servicio no está cableado (p.ej. la BD no tiene la
		// migración 012), devolvemos 404 limpio en vez de 500.
		a.NotFound(w, r)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/alojamiento/")
	if idStr == "" || strings.Contains(idStr, "/") {
		a.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		a.NotFound(w, r)
		return
	}
	aloja, err := a.AlojamientoSvc.Obtener(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.NotFound(w, r)
			return
		}
		log.Printf("ERROR obtener alojamiento %d: %v", id, err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	u := a.UsuarioActual(r)
	render(w, r, a.Plantillas, "alojamiento", datosAlojamiento{
		Titulo:      aloja.Nombre + " - SnowBreak",
		Descripcion: "Reserva noches en " + aloja.Nombre + ", cerca de la estación de " + aloja.NombreEstacion + ".",
		Activa:      "estaciones",
		Alojamiento: aloja,
		Usuario:     u,
		YaLogueado:  u != nil,
	})
}

// ApiReservarAlojamiento responde a POST /api/alojamientos/reservar.
//
// Body JSON:
//   {
//     "alojamiento_id":  12,
//     "fecha_entrada":   "2026-12-15",
//     "fecha_salida":    "2026-12-18",
//     "huespedes":       2
//   }
//
// Respuestas:
//   200 → ResumenReserva (id, noches, total, etc.)
//   400 → datos inválidos (fechas mal, huéspedes fuera de rango, etc.)
//   401 → sin sesión
//   404 → alojamiento no existe
//   500 → error interno
//
// Defensa en profundidad:
//   - El total se recalcula en SERVIDOR (precio_noche × noches). El
//     cliente nunca lo envía; aunque lo enviara, lo ignoramos.
//   - CSRF: el middleware general ya valida el token "double-submit".
//   - Rate limit: este endpoint se monta con rlEscritura en main.go.
func (a *App) ApiReservarAlojamiento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		escribirError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}
	if a.AlojamientoSvc == nil {
		escribirError(w, http.StatusServiceUnavailable, "servicio de alojamientos no disponible")
		return
	}

	u := a.UsuarioActual(r)
	if u == nil {
		escribirError(w, http.StatusUnauthorized, "inicia sesión para reservar")
		return
	}

	var p services.PeticionReserva
	if err := decodificarJSON(r, &p); err != nil {
		escribirError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}
	if p.AlojamientoID <= 0 {
		escribirError(w, http.StatusBadRequest, "alojamiento_id requerido")
		return
	}

	resumen, err := a.AlojamientoSvc.Reservar(r.Context(), u.ID, p)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrAlojamientoNoEncontrado):
			escribirError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, services.ErrFechaFormato),
			errors.Is(err, services.ErrFechaPasado),
			errors.Is(err, services.ErrFechaOrden),
			errors.Is(err, services.ErrFechaRangoMax),
			errors.Is(err, services.ErrHuespedes):
			escribirError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("ERROR reservar alojamiento: %v", err)
			escribirError(w, http.StatusInternalServerError, "error interno al reservar")
		}
		return
	}
	log.Printf("RESERVA alojamiento usuario=%d alojamiento=%d total=%.2f€", u.ID, p.AlojamientoID, resumen.TotalEur)
	escribirJSON(w, http.StatusOK, resumen)
}


// alojamientoJSON es la representacion publica de un alojamiento.
type alojamientoJSON struct {
	ID          int64   `json:"id"`
	EstacionID  int64   `json:"estacion_id"`
	Nombre      string  `json:"nombre"`
	Tipo        string  `json:"tipo"`
	TipoTexto   string  `json:"tipo_texto"`
	Imagen      string  `json:"imagen"`
	Distancia   float64 `json:"distancia_km"`
	Valoracion  float64 `json:"valoracion"`
	PrecioNoche float64 `json:"precio_noche"`
	Zona        string  `json:"zona"`
	Descripcion string  `json:"descripcion"`
}

// ApiAlojamientosEstacion responde a GET /api/alojamientos/estacion/{id}
// devolviendo los alojamientos asociados a esa estacion en JSON.
//
// Endpoint publico (sin autenticacion) — solo expone datos que ya estan
// en el catalogo, no info personal.
func (a *App) ApiAlojamientosEstacion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		escribirError(w, http.StatusMethodNotAllowed, "metodo no permitido")
		return
	}
	if a.AlojamientoSvc == nil {
		escribirJSON(w, http.StatusOK, []alojamientoJSON{})
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/alojamientos/estacion/")
	if idStr == "" || strings.Contains(idStr, "/") {
		escribirError(w, http.StatusBadRequest, "id de estacion requerido")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		escribirError(w, http.StatusBadRequest, "id de estacion invalido")
		return
	}
	lista, err := a.AlojamientoSvc.ListarPorEstacion(r.Context(), id)
	if err != nil {
		log.Printf("ERROR API alojamientos estacion=%d: %v", id, err)
		escribirError(w, http.StatusInternalServerError, "error interno")
		return
	}
	out := make([]alojamientoJSON, 0, len(lista))
	for _, l := range lista {
		out = append(out, alojamientoJSON{
			ID:          l.ID,
			EstacionID:  l.EstacionID,
			Nombre:      l.Nombre,
			Tipo:        l.Tipo,
			TipoTexto:   l.TipoTexto(),
			Imagen:      l.Imagen,
			Distancia:   l.Distancia,
			Valoracion:  l.Valoracion,
			PrecioNoche: l.PrecioNoche,
			Zona:        l.Zona,
			Descripcion: l.Descripcion,
		})
	}
	escribirJSON(w, http.StatusOK, out)
}
