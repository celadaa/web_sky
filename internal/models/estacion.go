package models

// Estacion representa una estación de esquí que aparece en el catálogo
// y en las fichas individuales. Todos los campos se leen/escriben desde
// la tabla `estaciones` de la base de datos SQLite.
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
}
