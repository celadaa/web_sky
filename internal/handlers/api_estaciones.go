// Package handlers — API REST de estaciones.
//
// Endpoint público (sin autenticación) que devuelve el listado de
// estaciones de esquí en JSON, **incluyendo las coordenadas geográficas
// (lat/lng)** que se usan en el visor de mapa de la página inicio.
//
//	GET /api/estaciones  →  [{id,nombre,ubicacion,lat,lng,...}, ...]
//
// La tabla `stations` de PostgreSQL no tiene columnas de coordenadas
// (mantenemos el esquema enfocado a datos meteo / operativos). En su
// lugar, este fichero guarda un mapa NOMBRE → (lat,lng) con los puntos
// reales conocidos de las estaciones sembradas. Si alguna estación
// nueva se añade en el futuro y todavía no tiene coordenadas, se
// devuelve con lat=0/lng=0 y se marca con `tiene_coords: false` para
// que el cliente la pueda filtrar y no acabe en mitad del Atlántico.
package handlers

import (
	"log"
	"net/http"
	"strings"
)

// coordsEstaciones — coordenadas reales (latitud, longitud) de las 27
// estaciones del catálogo. Se compara por nombre normalizado a minúsculas
// y sin tildes para tolerar pequeñas variaciones tipográficas.
var coordsEstaciones = map[string]struct {
	Lat float64
	Lng float64
}{
	// Pirineo Catalán
	"baqueira beret":  {42.7000, 0.9333},
	"la molina":       {42.3361, 1.9494},
	"masella":         {42.3744, 1.9261},
	"vall de nuria":   {42.3978, 2.1547},
	"vallter 2000":    {42.4231, 2.2722},
	"espot esqui":     {42.5917, 1.0911},
	"port aine":       {42.5572, 1.2156},
	"boi taull":       {42.5000, 0.8167},
	"tavascan":        {42.6519, 1.2914},
	"port del comte":  {42.2114, 1.5611},

	// Pirineo Aragonés
	"candanchu":  {42.7833, -0.5167},
	"astun":      {42.7889, -0.5028},
	"formigal":   {42.7811, -0.4108},
	"panticosa":  {42.7547, -0.2872},
	"cerler":     {42.5400, 0.4225},

	// Andorra
	"grandvalira":    {42.5392, 1.7308},
	"pal arinsal":    {42.5728, 1.4844},
	"ordino arcalis": {42.6450, 1.4836},
	"naturlandia":    {42.4639, 1.4711},

	// Sistema Penibético
	"sierra nevada": {37.0928, -3.3953},

	// Sistema Central
	"valdesqui":             {40.7833, -3.9000},
	"puerto de navacerrada": {40.7906, -4.0086},
	"la pinilla":            {41.3033, -3.4500},

	// Cordillera Cantábrica
	"alto campoo":        {43.0444, -4.4083},
	"san isidro":         {43.0419, -5.4178},
	"valgrande-pajares":  {43.0517, -5.7806},
	"leitariegos":        {43.0150, -6.4106},
}

// macizoDeEstaciones — clasificación de cada estación según el macizo
// montañoso al que pertenece. La clave es el nombre normalizado y el
// valor es una tupla con un código corto (para clases CSS y filtrado)
// y la etiqueta legible que se muestra al usuario.
var macizoDeEstaciones = map[string]struct {
	Clave  string
	Nombre string
}{
	// Pirineo Catalán
	"baqueira beret": {"pirineo-cat", "Pirineo Catalán"},
	"la molina":      {"pirineo-cat", "Pirineo Catalán"},
	"masella":        {"pirineo-cat", "Pirineo Catalán"},
	"vall de nuria":  {"pirineo-cat", "Pirineo Catalán"},
	"vallter 2000":   {"pirineo-cat", "Pirineo Catalán"},
	"espot esqui":    {"pirineo-cat", "Pirineo Catalán"},
	"port aine":      {"pirineo-cat", "Pirineo Catalán"},
	"boi taull":      {"pirineo-cat", "Pirineo Catalán"},
	"tavascan":       {"pirineo-cat", "Pirineo Catalán"},
	"port del comte": {"pirineo-cat", "Pirineo Catalán"},
	// Pirineo Aragonés
	"candanchu": {"pirineo-ara", "Pirineo Aragonés"},
	"astun":     {"pirineo-ara", "Pirineo Aragonés"},
	"formigal":  {"pirineo-ara", "Pirineo Aragonés"},
	"panticosa": {"pirineo-ara", "Pirineo Aragonés"},
	"cerler":    {"pirineo-ara", "Pirineo Aragonés"},
	// Andorra
	"grandvalira":    {"andorra", "Andorra"},
	"pal arinsal":    {"andorra", "Andorra"},
	"ordino arcalis": {"andorra", "Andorra"},
	"naturlandia":    {"andorra", "Andorra"},
	// Sistema Penibético
	"sierra nevada": {"penibetico", "Sierra Nevada"},
	// Sistema Central
	"valdesqui":             {"central", "Sistema Central"},
	"puerto de navacerrada": {"central", "Sistema Central"},
	"la pinilla":            {"central", "Sistema Central"},
	// Cordillera Cantábrica
	"alto campoo":       {"cantabrica", "Cordillera Cantábrica"},
	"san isidro":        {"cantabrica", "Cordillera Cantábrica"},
	"valgrande-pajares": {"cantabrica", "Cordillera Cantábrica"},
	"leitariegos":       {"cantabrica", "Cordillera Cantábrica"},
}

// estacionMapaJSON es la representación de cada estación en el JSON
// que consume el cliente de mapas (mapa.js).
type estacionMapaJSON struct {
	ID             int64   `json:"id"`
	Nombre         string  `json:"nombre"`
	Ubicacion      string  `json:"ubicacion"`
	Lat            float64 `json:"lat"`
	Lng            float64 `json:"lng"`
	TieneCoords    bool    `json:"tiene_coords"`
	Distancia      float64 `json:"distancia"`
	Altitud        string  `json:"altitud"`
	NieveBase      int     `json:"nieve_base"`
	NieveNueva     int     `json:"nieve_nueva"`
	Temperatura    int     `json:"temperatura"`
	PistasAbiertas int     `json:"pistas_abiertas"`
	PistasTotales  int     `json:"pistas_totales"`
	RemontesOp     int     `json:"remontes_op"`
	RemontesTot    int     `json:"remontes_tot"`
	Estado         string  `json:"estado"`        // excelente|buena|regular|cerrado
	EstadoTexto    string  `json:"estado_texto"`  // texto legible
	Macizo         string  `json:"macizo"`        // etiqueta legible (Pirineo Catalán, Sierra Nevada, …)
	MacizoClave    string  `json:"macizo_clave"`  // código corto para CSS/filtros
	Imagen         string  `json:"imagen"`
}

// calcularEstado deriva el estado operativo de la estación a partir
// del ratio pistas_abiertas / pistas_totales. Devuelve un código corto
// (que el cliente usa para colorear los marcadores) y un texto legible.
//
//	cerrado    → 0 pistas abiertas
//	regular    → < 50 % de las pistas abiertas
//	buena      → entre 50 % y 80 %
//	excelente  → >= 80 % de las pistas abiertas
func calcularEstado(abiertas, totales int) (string, string) {
	if totales <= 0 || abiertas <= 0 {
		return "cerrado", "Cerrada"
	}
	ratio := float64(abiertas) / float64(totales)
	switch {
	case ratio >= 0.8:
		return "excelente", "Excelente"
	case ratio >= 0.5:
		return "buena", "Buena"
	default:
		return "regular", "Regular"
	}
}

// ApiEstaciones responde a GET /api/estaciones devolviendo el listado
// completo en JSON con las coordenadas necesarias para pintar markers
// en el mapa. Es de SOLO LECTURA y pública: cualquier cliente puede
// consumirla sin estar autenticado.
func (a *App) ApiEstaciones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		escribirError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	// usuarioID = 0 → no marcamos favoritas (el mapa no las necesita).
	lista, err := a.EstacionSvc.Listar(r.Context(), 0)
	if err != nil {
		log.Printf("ERROR API listar estaciones: %v", err)
		escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}

	out := make([]estacionMapaJSON, 0, len(lista))
	for _, e := range lista {
		clave := normalizar(e.Nombre)
		coord, ok := coordsEstaciones[clave]
		estado, texto := calcularEstado(e.PistasAbiertas, e.PistasTotales)
		macizo, mOK := macizoDeEstaciones[clave]
		if !mOK {
			macizo.Clave = "otros"
			macizo.Nombre = "Otros"
		}
		out = append(out, estacionMapaJSON{
			ID:             e.ID,
			Nombre:         e.Nombre,
			Ubicacion:      e.Ubicacion,
			Lat:            coord.Lat,
			Lng:            coord.Lng,
			TieneCoords:    ok,
			Distancia:      e.Distancia,
			Altitud:        e.Altitud,
			NieveBase:      e.NieveBase,
			NieveNueva:     e.NieveNueva,
			Temperatura:    e.Temperatura,
			PistasAbiertas: e.PistasAbiertas,
			PistasTotales:  e.PistasTotales,
			RemontesOp:     e.RemontesOp,
			RemontesTot:    e.RemontesTot,
			Estado:         estado,
			EstadoTexto:    texto,
			Macizo:         macizo.Nombre,
			MacizoClave:    macizo.Clave,
			Imagen:         e.Imagen,
		})
	}
	escribirJSON(w, http.StatusOK, out)
}

// normalizar pasa a minúsculas y quita tildes para que la búsqueda
// en coordsEstaciones sea robusta ("Candanchú" → "candanchu").
func normalizar(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	r := strings.NewReplacer(
		"á", "a", "à", "a", "ä", "a", "â", "a",
		"é", "e", "è", "e", "ë", "e", "ê", "e",
		"í", "i", "ì", "i", "ï", "i", "î", "i",
		"ó", "o", "ò", "o", "ö", "o", "ô", "o",
		"ú", "u", "ù", "u", "ü", "u", "û", "u",
		"ñ", "n", "ç", "c",
	)
	return r.Replace(s)
}
