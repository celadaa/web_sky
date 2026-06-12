// Package handlers — páginas públicas y API del módulo de hoteles.
//
//	GET /hoteles                 → listado (filtro opcional ?estacion={id})
//	GET /hoteles/{slug}          → detalle con galería
//	GET /api/hoteles             → JSON (?station_id=&max_price=&max_distance=)
//
// La reserva no se procesa internamente: cada hotel enlaza a Booking o a su
// web oficial (campo validado en el service, solo https).
package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"skihub/internal/models"
	"skihub/internal/repository"
)

type datosHoteles struct {
	Titulo      string
	Descripcion string
	Activa      string
	Hoteles     []models.Hotel
	Estaciones  []models.Estacion
	// EstacionFiltro > 0 cuando el listado está filtrado por estación.
	EstacionFiltro int64
	EstacionNombre string
	Usuario        *models.Usuario
}

type datosHotel struct {
	Titulo      string
	Descripcion string
	Activa      string
	Hotel       *models.Hotel
	Usuario     *models.Usuario
}

// Hoteles responde a GET /hoteles con el listado de alojamientos,
// opcionalmente filtrado por estación (?estacion={id}).
func (a *App) Hoteles(w http.ResponseWriter, r *http.Request) {
	u := a.UsuarioActual(r)

	var filtro repository.FiltroHoteles
	if v := r.URL.Query().Get("estacion"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil && id > 0 {
			filtro.EstacionID = id
		}
	}

	lista, err := a.HotelSvc.Listar(r.Context(), filtro)
	if err != nil {
		log.Printf("ERROR listar hoteles: %v", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	// Listado de estaciones para el selector de filtro.
	estaciones, err := a.EstacionSvc.Listar(r.Context(), 0)
	if err != nil {
		log.Printf("ERROR listar estaciones para filtro hoteles: %v", err)
		estaciones = nil
	}

	nombreEstacion := ""
	if filtro.EstacionID > 0 {
		for i := range estaciones {
			if estaciones[i].ID == filtro.EstacionID {
				nombreEstacion = estaciones[i].Nombre
				break
			}
		}
	}

	titulo := "Hoteles cerca de las pistas - SnowBreak"
	if nombreEstacion != "" {
		titulo = "Hoteles en " + nombreEstacion + " - SnowBreak"
	}

	render(w, r, a.Plantillas, "hoteles", datosHoteles{
		Titulo:         titulo,
		Descripcion:    "Compara alojamientos cerca de las estaciones de esquí: precio, distancia a pistas, servicios y valoraciones.",
		Activa:         "hoteles",
		Hoteles:        lista,
		Estaciones:     estaciones,
		EstacionFiltro: filtro.EstacionID,
		EstacionNombre: nombreEstacion,
		Usuario:        u,
	})
}

// HotelDetalle responde a GET /hoteles/{slug} con la ficha y galería.
func (a *App) HotelDetalle(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/hoteles/")
	if slug == "" || strings.Contains(slug, "/") {
		a.NotFound(w, r)
		return
	}

	h, err := a.HotelSvc.ObtenerPorSlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.NotFound(w, r)
			return
		}
		log.Printf("ERROR obtener hotel %q: %v", slug, err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	render(w, r, a.Plantillas, "hotel", datosHotel{
		Titulo:      h.Nombre + " - Hoteles - SnowBreak",
		Descripcion: "Información, fotos, servicios y distancia a pistas de " + h.Nombre + ".",
		Activa:      "hoteles",
		Hotel:       h,
		Usuario:     a.UsuarioActual(r),
	})
}

// hotelJSON es la representación de un hotel en la API pública. La consume
// el planificador de estancia (paso de alojamiento).
type hotelJSON struct {
	ID                int64    `json:"id"`
	EstacionID        int64    `json:"estacion_id"`
	Nombre            string   `json:"nombre"`
	Slug              string   `json:"slug"`
	Ciudad            string   `json:"ciudad"`
	DistanciaEstacion float64  `json:"distancia_estacion_km"`
	PrecioDesde       float64  `json:"precio_desde"`
	Divisa            string   `json:"divisa"`
	Rating            float64  `json:"rating"`
	NumResenas        int      `json:"num_resenas"`
	Amenities         []string `json:"amenities"`
	Imagen            string   `json:"imagen"`
	EnlaceReserva     string   `json:"enlace_reserva"`
	URL               string   `json:"url"`
}

// ApiHoteles responde a GET /api/hoteles con el listado en JSON.
// Filtros opcionales: station_id, max_price, max_distance. Solo lectura
// y pública, igual que /api/estaciones.
func (a *App) ApiHoteles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		escribirError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	var filtro repository.FiltroHoteles
	q := r.URL.Query()
	if v := q.Get("station_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id < 0 {
			escribirError(w, http.StatusBadRequest, "station_id inválido")
			return
		}
		filtro.EstacionID = id
	}
	if v := q.Get("max_price"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f < 0 {
			escribirError(w, http.StatusBadRequest, "max_price inválido")
			return
		}
		filtro.PrecioMax = f
	}
	if v := q.Get("max_distance"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f < 0 {
			escribirError(w, http.StatusBadRequest, "max_distance inválido")
			return
		}
		filtro.DistanciaMax = f
	}

	lista, err := a.HotelSvc.Listar(r.Context(), filtro)
	if err != nil {
		log.Printf("ERROR API listar hoteles: %v", err)
		escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}

	out := make([]hotelJSON, 0, len(lista))
	for _, h := range lista {
		out = append(out, hotelJSON{
			ID:                h.ID,
			EstacionID:        h.EstacionID,
			Nombre:            h.Nombre,
			Slug:              h.Slug,
			Ciudad:            h.Ciudad,
			DistanciaEstacion: h.DistanciaEstacion,
			PrecioDesde:       h.PrecioDesde,
			Divisa:            h.Divisa,
			Rating:            h.Rating,
			NumResenas:        h.NumResenas,
			Amenities:         h.ListaAmenities(),
			Imagen:            h.ImagenURL,
			EnlaceReserva:     h.EnlaceReserva(),
			URL:               "/hoteles/" + h.Slug,
		})
	}
	escribirJSON(w, http.StatusOK, out)
}
