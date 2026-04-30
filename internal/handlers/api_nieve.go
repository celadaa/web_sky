// Package handlers — API REST de pistas en directo.
//
// Estos endpoints actúan como proxy seguro entre el navegador y la
// fuente de datos (infonieve.es). Aplican cache, evitan exponer la
// URL/lógica del scraping al cliente y devuelven JSON con un formato
// estable. Todos son de SOLO LECTURA y públicos.
//
//	GET /api/nieve/estaciones?lat=X&lng=Y[&limit=N]
//	    → listado ordenado por distancia desde (X,Y).
//	GET /api/nieve/estaciones/{slug}
//	    → ficha completa (pistas, remontes, webcams).
//	GET /api/nieve/regiones
//	    → listado de regiones soportadas.
package handlers

import (
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"skihub/internal/infonieve"
	"skihub/internal/services"
)

var reSlugValido = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{0,60}$`)

// ApiNieveEstaciones — listado en directo ordenado por distancia.
func (a *App) ApiNieveEstaciones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		escribirError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}
	if a.NieveSvc == nil {
		escribirError(w, http.StatusServiceUnavailable, "servicio de nieve no disponible")
		return
	}

	lat, lng, err := parsearLatLng(r.URL.Query())
	if err != nil {
		escribirError(w, http.StatusBadRequest, err.Error())
		return
	}
	limite := parsearLimite(r.URL.Query().Get("limit"))

	lista, err := a.NieveSvc.EstacionesCercanas(lat, lng, limite)
	if err != nil {
		escribirErrorScraping(w, err, "listado de estaciones")
		return
	}
	escribirJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"total": len(lista),
		"data":  lista,
	})
}

// ApiNieveEstacion — detalle por slug.
func (a *App) ApiNieveEstacion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		escribirError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}
	if a.NieveSvc == nil {
		escribirError(w, http.StatusServiceUnavailable, "servicio de nieve no disponible")
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/api/nieve/estaciones/")
	slug = strings.TrimSuffix(slug, "/")
	if !reSlugValido.MatchString(slug) {
		escribirError(w, http.StatusBadRequest, "slug inválido")
		return
	}

	det, err := a.NieveSvc.Detalle(slug)
	if err != nil {
		escribirErrorScraping(w, err, "estación "+slug)
		return
	}
	escribirJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"data": det,
	})
}

// ApiNieveRegiones — lista de regiones.
func (a *App) ApiNieveRegiones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		escribirError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}
	if a.NieveSvc == nil {
		escribirError(w, http.StatusServiceUnavailable, "servicio de nieve no disponible")
		return
	}
	escribirJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"data": a.NieveSvc.Regiones(),
	})
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// parsearLatLng acepta "lat" y "lng" como query strings; si no se pasan,
// devuelve 0,0 (sin orden por distancia). Si se pasan parcialmente o
// fuera de rango, devuelve error.
func parsearLatLng(q map[string][]string) (float64, float64, error) {
	latStr, lngStr := primero(q, "lat"), primero(q, "lng")
	if latStr == "" && lngStr == "" {
		return 0, 0, nil
	}
	if latStr == "" || lngStr == "" {
		return 0, 0, errors.New("lat y lng deben ir juntos")
	}
	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	if err1 != nil || err2 != nil {
		return 0, 0, errors.New("lat/lng no numéricos")
	}
	if !services.CoordsValidas(lat, lng) {
		return 0, 0, services.ErrCoordenadasInvalidas
	}
	return lat, lng, nil
}

func primero(q map[string][]string, k string) string {
	if v, ok := q[k]; ok && len(v) > 0 {
		return strings.TrimSpace(v[0])
	}
	return ""
}

func parsearLimite(s string) int {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 || n > 200 {
		return 0
	}
	return n
}

// escribirErrorScraping mapea los errores del paquete infonieve a códigos
// HTTP y a mensajes amigables en español.
func escribirErrorScraping(w http.ResponseWriter, err error, contexto string) {
	switch {
	case errors.Is(err, infonieve.ErrNotFound):
		escribirError(w, http.StatusNotFound, contexto+" no encontrada")
	case errors.Is(err, infonieve.ErrRateLimited):
		escribirError(w, http.StatusTooManyRequests,
			"infonieve.es está limitando las peticiones, prueba en unos minutos")
	case errors.Is(err, infonieve.ErrTimeout):
		escribirError(w, http.StatusGatewayTimeout, "tiempo de espera agotado")
	default:
		log.Printf("ERROR scraping %s: %v", contexto, err)
		escribirError(w, http.StatusBadGateway,
			"no se pudo obtener "+contexto+" desde la fuente externa")
	}
}
