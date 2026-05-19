package models

import "time"

// Alojamiento representa un hotel/hostal/apartamento cercano a una
// estación de esquí. Se lee de la tabla `lodgings`. Convención del
// proyecto: SQL en inglés, Go en español.
type Alojamiento struct {
	ID            int64
	EstacionID    int64
	Nombre        string
	Tipo          string  // hotel | hostal | apartamento | albergue | casa_rural
	Imagen        string  // URL absoluta (Unsplash) o ruta /static/
	Distancia     float64 // km a la estación o pistas
	Valoracion    float64 // 0.0 a 5.0
	PrecioNoche   float64 // EUR
	Zona          string  // p.ej. "A pie de pistas", "Centro de Pradollano"
	Descripcion   string
	CreadoEn      time.Time
	ActualizadoEn time.Time

	// Campos enriquecidos no persistidos (los rellena el servicio).
	NombreEstacion string
}

// TipoTexto devuelve el tipo en formato legible para la UI.
func (a Alojamiento) TipoTexto() string {
	switch a.Tipo {
	case "hotel":
		return "Hotel"
	case "hostal":
		return "Hostal"
	case "apartamento":
		return "Apartamento"
	case "albergue":
		return "Albergue"
	case "casa_rural":
		return "Casa rural"
	default:
		return "Alojamiento"
	}
}

// ReservaAlojamiento es una reserva confirmada (o simulada) por un
// usuario autenticado. Se persiste en `lodging_bookings` con los datos
// del alojamiento congelados como snapshot histórico.
type ReservaAlojamiento struct {
	ID             int64
	UsuarioID      int64
	AlojamientoID  int64
	NombreAloja    string
	TipoAloja      string
	EstacionID     int64
	NombreEstacion string
	FechaEntrada   time.Time
	FechaSalida    time.Time
	Noches         int
	Huespedes      int
	PrecioNoche    float64
	TotalEur       float64
	Estado         string // confirmed | cancelled
	CreadoEn       time.Time
}
