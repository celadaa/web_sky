// Package handlers — API REST de estaciones.
//
// Endpoint público (sin autenticación) que devuelve el listado de
// estaciones de esquí en JSON, **incluyendo las coordenadas geográficas
// (lat/lng)** que se usan en el visor de mapa de la página inicio.
//
//	GET /api/estaciones  →  [{id,nombre,ubicacion,lat,lng,...}, ...]
//
// La tabla `estaciones` de SQLite no tiene columnas de coordenadas (y
// así se mantiene el esquema intacto). En su lugar, este fichero
// guarda un mapa NOMBRE → (lat,lng) con los puntos reales conocidos
// de las estaciones sembradas. Si alguna estación nueva se añade en
// el futuro y todavía no tiene coordenadas, se devuelve con lat=0/lng=0
// y se marca con `tiene_coords: false` para que el cliente la pueda
// filtrar y no acabe en mitad del Atlántico.
package handlers

import (
	"log"
	"net/http"
	"strings"
)

// coordsEstaciones — coordenadas reales (latitud, longitud) de las
// estaciones del seed. Se compara por nombre normalizado a minúsculas
// y sin tildes para tolerar pequeñas variaciones tipográficas.
var coordsEstaciones = map[string]struct {
	Lat float64
	Lng float64
}{
	"candanchu":      {42.7833, -0.5167},
	"sierra nevada":  {37.0928, -3.3953},
	"grandvalira":    {42.5392, 1.7308},
	"formigal":       {42.7811, -0.4108},
	"baqueira beret": {42.7000, 0.9333},
	"cerler":         {42.5400, 0.4225},
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
	PistasAbiertas int     `json:"pistas_abiertas"`
	PistasTotales  int     `json:"pistas_totales"`
	Imagen         string  `json:"imagen"`
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
	lista, err := a.EstacionSvc.Listar(0)
	if err != nil {
		log.Printf("ERROR API listar estaciones: %v", err)
		escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}

	out := make([]estacionMapaJSON, 0, len(lista))
	for _, e := range lista {
		clave := normalizar(e.Nombre)
		coord, ok := coordsEstaciones[clave]
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
			PistasAbiertas: e.PistasAbiertas,
			PistasTotales:  e.PistasTotales,
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
