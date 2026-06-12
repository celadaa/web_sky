package services

import (
	"testing"

	"skihub/internal/models"
)

func TestSanearURLExterna(t *testing.T) {
	casos := []struct {
		in     string
		want   string
		quiero bool
	}{
		{"", "", true},
		{"   ", "", true},
		{"https://www.booking.com/hotel/es/demo.html", "https://www.booking.com/hotel/es/demo.html", true},
		{"https://images.unsplash.com/photo-1?q=80", "https://images.unsplash.com/photo-1?q=80", true},
		{"/static/news-img/foto.jpg", "/static/news-img/foto.jpg", true},
		{"http://inseguro.com", "", false},
		{"javascript:alert(1)", "", false},
		{"JAVASCRIPT:alert(1)", "", false},
		{"data:text/html;base64,xx", "", false},
		{"ftp://servidor/fichero", "", false},
		{"https://", "", false},
		{"//evil.com/x.jpg", "", false},
		{"relativa/sin/static", "", false},
	}
	for _, c := range casos {
		got, ok := SanearURLExterna(c.in)
		if ok != c.quiero || got != c.want {
			t.Errorf("SanearURLExterna(%q) = (%q, %v); esperaba (%q, %v)",
				c.in, got, ok, c.want, c.quiero)
		}
	}
}

func TestGenerarSlug(t *testing.T) {
	casos := []struct {
		in, want string
	}{
		{"Hotel Demo Pie de Pistas", "hotel-demo-pie-de-pistas"},
		{"Hôtel Pic d'Or 4★", "hotel-pic-d-or-4"},
		{"  Cabaña / Montaña  ", "cabana-montana"},
		{"---", ""},
		{"ÁÉÍÓÚ ñç", "aeiou-nc"},
	}
	for _, c := range casos {
		if got := GenerarSlug(c.in); got != c.want {
			t.Errorf("GenerarSlug(%q) = %q; esperaba %q", c.in, got, c.want)
		}
	}
}

func TestValidarHotel(t *testing.T) {
	s := &HotelService{}

	h := nuevoHotelValido()
	if err := s.validar(h); err != nil {
		t.Fatalf("hotel válido rechazado: %v", err)
	}
	if h.Divisa != "EUR" {
		t.Errorf("la divisa por defecto debería ser EUR, fue %q", h.Divisa)
	}

	h = nuevoHotelValido()
	h.Nombre = "  "
	if err := s.validar(h); err == nil {
		t.Error("nombre vacío debería fallar")
	}

	h = nuevoHotelValido()
	h.EstacionID = 0
	if err := s.validar(h); err == nil {
		t.Error("estación 0 debería fallar")
	}

	h = nuevoHotelValido()
	h.Rating = 5.5
	if err := s.validar(h); err == nil {
		t.Error("rating > 5 debería fallar")
	}

	h = nuevoHotelValido()
	h.BookingURL = "javascript:alert(1)"
	if err := s.validar(h); err == nil {
		t.Error("booking_url javascript: debería fallar")
	}

	h = nuevoHotelValido()
	h.PrecioDesde = -1
	if err := s.validar(h); err == nil {
		t.Error("precio negativo debería fallar")
	}
}

func nuevoHotelValido() *models.Hotel {
	return &models.Hotel{
		Nombre:      "Hotel Test",
		EstacionID:  1,
		Rating:      4.2,
		PrecioDesde: 80,
		BookingURL:  "https://www.booking.com/hotel/es/test.html",
	}
}
