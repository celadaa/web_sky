package handlers

import (
	"bytes"
	"testing"
	"time"

	"skihub/internal/models"
)

// Smoke test: las plantillas del módulo de hoteles renderizan sin error
// con datos representativos (con y sin imágenes/enlaces/rating).
func TestRenderPlantillasHoteles(t *testing.T) {
	cache, err := CargarPlantillas("../../web/templates")
	if err != nil {
		t.Fatalf("cargando plantillas: %v", err)
	}

	conTodo := models.Hotel{
		ID: 1, EstacionID: 2, Nombre: "Hotel Test", Slug: "hotel-test",
		Descripcion: "Descripción", Direccion: "Calle 1", Ciudad: "Vielha", Pais: "España",
		DistanciaEstacion: 1.5, PrecioDesde: 120, Divisa: "EUR",
		Rating: 4.5, NumResenas: 100,
		Amenities:  "wifi,spa,parking",
		BookingURL: "https://www.booking.com/x", OficialURL: "https://hotel.example",
		ImagenURL: "https://images.unsplash.com/a.jpg",
		CreadoEn:  time.Now(), Actualizado: time.Now(),
		EstacionNombre: "Baqueira Beret",
		Imagenes: []models.HotelImagen{
			{ID: 1, HotelID: 1, URL: "https://images.unsplash.com/a.jpg", AltText: "Foto 1", EsPrimaria: true},
			{ID: 2, HotelID: 1, URL: "https://images.unsplash.com/b.jpg", AltText: "Foto 2"},
		},
	}
	sinNada := models.Hotel{
		ID: 2, EstacionID: 2, Nombre: "Hostal Vacío", Slug: "hostal-vacio",
		EstacionNombre: "Sierra Nevada",
	}
	estaciones := []models.Estacion{{ID: 2, Nombre: "Baqueira Beret"}}

	casos := []struct {
		pagina string
		datos  any
	}{
		{"hoteles", datosHoteles{Titulo: "t", Hoteles: []models.Hotel{conTodo, sinNada}, Estaciones: estaciones, EstacionFiltro: 2, EstacionNombre: "Baqueira Beret"}},
		{"hoteles", datosHoteles{Titulo: "t", Hoteles: nil, Estaciones: estaciones}},
		{"hotel", datosHotel{Titulo: "t", Hotel: &conTodo}},
		{"hotel", datosHotel{Titulo: "t", Hotel: &sinNada}},
		{"admin_hoteles", datosAdminHoteles{Titulo: "t", Hoteles: []models.Hotel{conTodo}, Usuario: &models.Usuario{Nombre: "Admin"}}},
		{"admin_hotel", datosAdminHotel{Titulo: "t", Detalle: &conTodo, Estaciones: estaciones, Usuario: &models.Usuario{Nombre: "Admin"}}},
		{"admin_hotel", datosAdminHotel{Titulo: "t", Detalle: &models.Hotel{Divisa: "EUR"}, Estaciones: estaciones, EsNuevo: true, Usuario: &models.Usuario{Nombre: "Admin"}}},
	}
	for i, c := range casos {
		var buf bytes.Buffer
		if err := cache[c.pagina].ExecuteTemplate(&buf, "layout", c.datos); err != nil {
			t.Errorf("caso %d (%s): %v", i, c.pagina, err)
		}
	}
}
