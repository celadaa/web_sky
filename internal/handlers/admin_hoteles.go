// Package handlers — panel de administración de hoteles.
//
//	GET  /admin/hoteles          → listado
//	GET  /admin/hotel/           → formulario de alta
//	GET  /admin/hotel/{id}       → formulario de edición
//	POST /admin/hoteles/guardar  → crear/actualizar (CSRF + admin)
//	POST /admin/hoteles/borrar   → eliminar (CSRF + admin)
//
// Las imágenes se introducen como una URL por línea (https o /static/...);
// la primera línea pasa a ser la imagen primaria de la galería.
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
	"skihub/internal/services"
)

type datosAdminHoteles struct {
	Titulo      string
	Descripcion string
	Activa      string
	Hoteles     []models.Hotel
	Total       int
	Mensaje     string
	Error       string
	Usuario     *models.Usuario
}

type datosAdminHotel struct {
	Titulo      string
	Descripcion string
	Activa      string
	Detalle     *models.Hotel
	Estaciones  []models.Estacion
	// GaleriaTexto: URLs de la galería, una por línea, para el textarea.
	GaleriaTexto string
	EsNuevo      bool
	Error        string
	Usuario      *models.Usuario
}

// AdminHoteles muestra en /admin/hoteles el listado completo.
func (a *App) AdminHoteles(w http.ResponseWriter, r *http.Request) {
	actual := a.requerirAdmin(w, r)
	if actual == nil {
		return
	}
	lista, err := a.HotelSvc.Listar(r.Context(), repository.FiltroHoteles{})
	if err != nil {
		log.Printf("ERROR admin listar hoteles: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}
	render(w, r, a.Plantillas, "admin_hoteles", datosAdminHoteles{
		Titulo:      "Administración - Hoteles",
		Descripcion: "Panel de administración de hoteles de SnowBreak.",
		Activa:      "admin",
		Hoteles:     lista,
		Total:       len(lista),
		Mensaje:     r.URL.Query().Get("msg"),
		Error:       r.URL.Query().Get("err"),
		Usuario:     actual,
	})
}

// AdminHotelForm muestra el formulario de alta (/admin/hotel/) o de
// edición (/admin/hotel/{id}).
func (a *App) AdminHotelForm(w http.ResponseWriter, r *http.Request) {
	actual := a.requerirAdmin(w, r)
	if actual == nil {
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/admin/hotel/")
	if strings.Contains(idStr, "/") {
		a.NotFound(w, r)
		return
	}

	datos := datosAdminHotel{
		Titulo:      "Administración - Nuevo hotel",
		Descripcion: "Alta de hotel en SnowBreak.",
		Activa:      "admin",
		Detalle:     &models.Hotel{Divisa: "EUR", Pais: "España"},
		EsNuevo:     true,
		Error:       r.URL.Query().Get("err"),
		Usuario:     actual,
	}

	if idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			a.NotFound(w, r)
			return
		}
		h, err := a.HotelSvc.ObtenerPorID(r.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.NotFound(w, r)
				return
			}
			log.Printf("ERROR admin obtener hotel %d: %v", id, err)
			http.Error(w, "error interno del servidor", http.StatusInternalServerError)
			return
		}
		urls := make([]string, 0, len(h.Imagenes))
		for _, img := range h.Imagenes {
			urls = append(urls, img.URL)
		}
		datos.Titulo = "Administración - " + h.Nombre
		datos.Detalle = h
		datos.GaleriaTexto = strings.Join(urls, "\n")
		datos.EsNuevo = false
	}

	estaciones, err := a.EstacionSvc.Listar(r.Context(), 0)
	if err != nil {
		log.Printf("ERROR admin listar estaciones: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}
	datos.Estaciones = estaciones

	render(w, r, a.Plantillas, "admin_hotel", datos)
}

// AdminGuardarHotel procesa el alta/edición (POST /admin/hoteles/guardar).
func (a *App) AdminGuardarHotel(w http.ResponseWriter, r *http.Request) {
	actual := a.requerirAdmin(w, r)
	if actual == nil {
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/hoteles?err="+urlEncode("Formulario inválido"), http.StatusSeeOther)
		return
	}

	id, _ := strconv.ParseInt(r.FormValue("hotel_id"), 10, 64)
	estacionID, _ := strconv.ParseInt(r.FormValue("estacion_id"), 10, 64)
	precio, _ := strconv.ParseFloat(strings.ReplaceAll(r.FormValue("precio_desde"), ",", "."), 64)
	distancia, _ := strconv.ParseFloat(strings.ReplaceAll(r.FormValue("distancia"), ",", "."), 64)
	rating, _ := strconv.ParseFloat(strings.ReplaceAll(r.FormValue("rating"), ",", "."), 64)
	resenas, _ := strconv.Atoi(r.FormValue("num_resenas"))

	h := &models.Hotel{
		ID:                id,
		EstacionID:        estacionID,
		Nombre:            strings.TrimSpace(r.FormValue("nombre")),
		Slug:              "", // se regenera/conserva en el service
		Descripcion:       strings.TrimSpace(r.FormValue("descripcion")),
		Direccion:         strings.TrimSpace(r.FormValue("direccion")),
		Ciudad:            strings.TrimSpace(r.FormValue("ciudad")),
		Pais:              strings.TrimSpace(r.FormValue("pais")),
		DistanciaEstacion: distancia,
		PrecioDesde:       precio,
		Divisa:            strings.TrimSpace(r.FormValue("divisa")),
		Rating:            rating,
		NumResenas:        resenas,
		Amenities:         strings.TrimSpace(r.FormValue("amenities")),
		BookingURL:        r.FormValue("booking_url"),
		OficialURL:        r.FormValue("official_url"),
		ImagenURL:         r.FormValue("main_image_url"),
	}
	// Al editar conservamos el slug existente (URLs estables).
	if id > 0 {
		if prev, err := a.HotelSvc.ObtenerPorID(r.Context(), id); err == nil {
			h.Slug = prev.Slug
		}
	}

	galeria := strings.Split(strings.ReplaceAll(r.FormValue("galeria"), "\r\n", "\n"), "\n")

	nuevoID, err := a.HotelSvc.Guardar(r.Context(), h, galeria)
	if err != nil {
		log.Printf("ERROR admin guardar hotel (admin %d): %v", actual.ID, err)
		destino := "/admin/hotel/"
		if id > 0 {
			destino += strconv.FormatInt(id, 10)
		}
		http.Redirect(w, r, destino+"?err="+urlEncode(mensajeErrorHotel(err)), http.StatusSeeOther)
		return
	}
	log.Printf("Admin %d guardó hotel %d (%s)", actual.ID, nuevoID, h.Nombre)
	http.Redirect(w, r, "/admin/hoteles?msg="+urlEncode("Hotel guardado"), http.StatusSeeOther)
}

// AdminBorrarHotel elimina un hotel (POST /admin/hoteles/borrar).
func (a *App) AdminBorrarHotel(w http.ResponseWriter, r *http.Request) {
	actual := a.requerirAdmin(w, r)
	if actual == nil {
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	id, ok := parseIDForm(r, "hotel_id")
	if !ok {
		http.Redirect(w, r, "/admin/hoteles?err=ID+inv%C3%A1lido", http.StatusSeeOther)
		return
	}
	if err := a.HotelSvc.Borrar(r.Context(), id); err != nil {
		log.Printf("ERROR admin borrar hotel %d: %v", id, err)
		http.Redirect(w, r, "/admin/hoteles?err="+urlEncode("Error al borrar"), http.StatusSeeOther)
		return
	}
	log.Printf("Admin %d borró el hotel %d", actual.ID, id)
	http.Redirect(w, r, "/admin/hoteles?msg="+urlEncode("Hotel eliminado"), http.StatusSeeOther)
}

// mensajeErrorHotel traduce los errores de validación del service a
// mensajes amigables; cualquier otro error se generaliza.
func mensajeErrorHotel(err error) string {
	switch {
	case errors.Is(err, services.ErrHotelNombreVacio),
		errors.Is(err, services.ErrHotelSinEstacion),
		errors.Is(err, services.ErrHotelRatingInvalid),
		errors.Is(err, services.ErrHotelPrecioInvalid),
		errors.Is(err, services.ErrHotelURLInvalida):
		return err.Error()
	default:
		return "Error al guardar el hotel"
	}
}
