package models

import (
	"strings"
	"time"
)

// Hotel representa un alojamiento asociado a una estación de esquí.
// Se persiste en la tabla `hotels` (columnas SQL en inglés, identificadores
// Go en español, igual que Estacion). La reserva no se procesa internamente:
// EnlaceReserva() devuelve el enlace externo donde se completa.
type Hotel struct {
	ID          int64
	EstacionID  int64
	Nombre      string
	Slug        string
	Descripcion string

	Direccion string
	Ciudad    string
	Pais      string
	// DistanciaEstacion en km hasta la estación asociada.
	DistanciaEstacion float64

	// PrecioDesde es orientativo, por noche, en la divisa indicada.
	PrecioDesde float64
	Divisa      string
	Rating      float64
	NumResenas  int

	// Amenities se guarda en BD como texto separado por comas;
	// usar ListaAmenities() para iterarlo en plantillas.
	Amenities string

	BookingURL  string
	OficialURL  string
	ImagenURL   string
	CreadoEn    time.Time
	Actualizado time.Time

	// EstacionNombre lo rellena el service con un JOIN; no se persiste.
	EstacionNombre string

	// Imagenes de la galería; las rellena el service solo en la vista detalle.
	Imagenes []HotelImagen
}

// HotelImagen es una foto de la galería de un hotel (tabla `hotel_images`).
type HotelImagen struct {
	ID         int64
	HotelID    int64
	URL        string
	AltText    string
	FuenteNom  string
	FuenteURL  string
	Orden      int
	EsPrimaria bool
	CreadoEn   time.Time
}

// ListaAmenities trocea el campo Amenities por comas, limpiando espacios
// y descartando vacíos, listo para iterar con range en las plantillas.
func (h Hotel) ListaAmenities() []string {
	if strings.TrimSpace(h.Amenities) == "" {
		return nil
	}
	partes := strings.Split(h.Amenities, ",")
	out := make([]string, 0, len(partes))
	for _, p := range partes {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// EnlaceReserva devuelve el enlace externo preferente para reservar:
// Booking si existe, si no la web oficial, si no cadena vacía (la
// plantilla oculta el botón en ese caso).
func (h Hotel) EnlaceReserva() string {
	if h.BookingURL != "" {
		return h.BookingURL
	}
	return h.OficialURL
}

// TieneRating indica si hay valoración que mostrar (rating > 0).
func (h Hotel) TieneRating() bool {
	return h.Rating > 0
}
