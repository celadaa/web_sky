package models

import "time"

// Estacion representa una estación de esquí que aparece en el catálogo
// y en las fichas individuales. Todos los campos se leen/escriben desde
// la tabla `stations` de PostgreSQL (los nombres de columna en SQL están
// en inglés; en Go conservamos los identificadores en español).
type Estacion struct {
	ID             int64
	Nombre         string
	Ubicacion      string
	Distancia      float64 // km desde la ubicación del usuario (demo: Madrid)
	Temperatura    int     // °C
	NieveBase      int     // cm
	NieveNueva     int     // cm caídos en las últimas 24 h
	PistasAbiertas int
	PistasTotales  int
	RemontesOp     int
	RemontesTot    int
	UltimaNevada   string
	Altitud        string
	KmEsquiables   float64
	Dificultad     string
	Telefono       string
	Imagen         string
	Descripcion    string

	// Precios del forfait (por día) según el tipo de pase. Se usan en el
	// widget de compra y en la página de listado de forfaits.
	PrecioAdulto float64
	PrecioNino   float64
	PrecioSenior float64

	// EsFavorita no se persiste en la tabla estaciones; lo rellena el
	// servicio en función del usuario autenticado que vea la página.
	EsFavorita bool

	// Campos dinámicos del "parte de nieve". Los rellena EstacionService
	// con datos en tiempo real de infonieve.es a través de NieveService.
	// NO se leen de la BD y NO se generan de forma ficticia: si no hay
	// datos reales disponibles quedan a cero/vacío y TieneDatosReales=false.
	NieveMin         int       // cm — espesor mínimo (= NieveMax cuando viene de infonieve)
	NieveMax         int       // cm — espesor máximo
	Viento           string    // vacío si no está disponible en la fuente
	ParteActualizado time.Time // momento de la última actualización real

	// TieneDatosReales es true cuando los campos dinámicos anteriores
	// provienen de infonieve.es en tiempo real. Úsalo en plantillas para
	// mostrar "—" en lugar de cero cuando no hay dato.
	TieneDatosReales bool
}

// ParteActualizadoTexto devuelve el momento del parte en formato
// dd/mm/aaaa HH:MM. Si no hay timestamp, devuelve cadena vacía.
func (e Estacion) ParteActualizadoTexto() string {
	if e.ParteActualizado.IsZero() {
		return ""
	}
	return e.ParteActualizado.Format("02/01/2006 15:04")
}

// ParteHora devuelve solo HH:MM, útil para el badge "Parte 12:45".
func (e Estacion) ParteHora() string {
	if e.ParteActualizado.IsZero() {
		return ""
	}
	return e.ParteActualizado.Format("15:04")
}

// ParteHaceMinutos devuelve los minutos transcurridos desde el último
// parte. Útil para construir frases tipo "hace X minutos" en la UI.
func (e Estacion) ParteHaceMinutos() int {
	if e.ParteActualizado.IsZero() {
		return -1
	}
	d := time.Since(e.ParteActualizado)
	if d < 0 {
		return 0
	}
	return int(d.Minutes())
}

// EstadoTexto devuelve "Abierta", "Apertura parcial" o "Cerrada"
// según las pistas abiertas. Para evitar marcar como "parcial" toda
// estación a la que le falte alguna pista, consideramos "Abierta"
// cualquier ratio >= 80 % (umbral típico en la industria del esquí
// para considerar una estación "100 % operativa").
func (e Estacion) EstadoTexto() string {
	if e.PistasAbiertas <= 0 {
		return "Cerrada"
	}
	if e.PistasTotales > 0 && e.PistasAbiertas*5 >= e.PistasTotales*4 {
		return "Abierta"
	}
	return "Apertura parcial"
}

// EstadoClase devuelve el modificador CSS a aplicar al badge de estado:
// "abierta", "parcial" o "cerrada". Se usa para mantener una sola fuente
// de verdad entre el texto y la clase visual.
func (e Estacion) EstadoClase() string {
	switch e.EstadoTexto() {
	case "Abierta":
		return "abierta"
	case "Cerrada":
		return "cerrada"
	default:
		return "parcial"
	}
}
