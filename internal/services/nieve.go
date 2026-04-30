// Package services — capa NieveService.
//
// NieveService añade tres capas sobre el scraper de infonieve.es:
//
//  1. Cache en memoria con TTL. Evita golpear la fuente en cada petición
//     (lista 10 min, detalle 5 min) y cumple con la nota de buen
//     ciudadano del README de infonieve-api.
//  2. Geolocalización. Asigna coordenadas conocidas por slug y calcula
//     distancia desde la posición del usuario con la fórmula de
//     Haversine.
//  3. DTO estable. Devuelve un objeto que el frontend consume tal cual,
//     listo para pintar tarjetas (campos predecibles aunque la fuente
//     omita datos).
//
// Las claves de error de scraping se pasan tal cual hacia arriba para
// que el handler pueda mapearlas a códigos HTTP.
package services

import (
	"errors"
	"math"
	"sort"
	"sync"
	"time"

	"skihub/internal/infonieve"
)

// EstacionDirecto es el DTO público que se devuelve al frontend.
// Combina datos del scraping con el cálculo de distancia.
type EstacionDirecto struct {
	Slug         string             `json:"slug"`
	Nombre       string             `json:"nombre"`
	URL          string             `json:"url"`
	Estado       infonieve.Estado   `json:"estado,omitempty"`
	Lat          float64            `json:"lat"`
	Lng          float64            `json:"lng"`
	TieneCoords  bool               `json:"tiene_coords"`
	DistanciaKm  *float64           `json:"distancia_km,omitempty"`
	Remontes     infonieve.Fraccion `json:"remontes"`
	Pistas       infonieve.Fraccion `json:"pistas"`
	Kilometros   infonieve.Fraccion `json:"kilometros"`
	NieveCm      *float64           `json:"nieve_cm,omitempty"`
	CalidadNieve string             `json:"calidad_nieve,omitempty"`
	Temperatura  string             `json:"temperatura,omitempty"`
}

// NieveService orquesta scraper + cache.
type NieveService struct {
	cli         *infonieve.Client
	mu          sync.RWMutex
	cache       map[string]cacheEntry
	listaTTL    time.Duration
	detalleTTL  time.Duration
	maxResults  int
}

type cacheEntry struct {
	dato    any
	expira  time.Time
}

// NuevoNieveService construye el servicio con TTLs sensatos.
func NuevoNieveService() *NieveService {
	return &NieveService{
		cli:        infonieve.NewClient(),
		cache:      make(map[string]cacheEntry),
		listaTTL:   10 * time.Minute,
		detalleTTL: 5 * time.Minute,
		maxResults: 200,
	}
}

// EstacionesCercanas devuelve el listado completo, ordenado por distancia
// si lat/lng son válidos. Si la posición es 0/0 no se calcula distancia
// y se devuelven en el orden recibido del scraping.
//
// El parámetro limite (>0) acota el número de resultados; pasar 0 los
// devuelve todos.
func (s *NieveService) EstacionesCercanas(lat, lng float64, limite int) ([]EstacionDirecto, error) {
	lista, err := s.listadoCacheado()
	if err != nil {
		return nil, err
	}

	out := make([]EstacionDirecto, 0, len(lista))
	hayUbicacion := lat != 0 || lng != 0

	for _, e := range lista {
		c, ok := infonieve.CoordPorSlug(e.Slug)
		dto := EstacionDirecto{
			Slug:         e.Slug,
			Nombre:       e.Nombre,
			URL:          e.URL,
			Estado:       e.Estado,
			Lat:          c.Lat,
			Lng:          c.Lng,
			TieneCoords:  ok,
			Remontes:     e.Remontes,
			Pistas:       e.Pistas,
			Kilometros:   e.Kilometros,
			NieveCm:      e.NieveCm,
			CalidadNieve: e.CalidadNieve,
			Temperatura:  e.Temperatura,
		}
		if ok && hayUbicacion {
			d := DistanciaHaversineKm(lat, lng, c.Lat, c.Lng)
			dto.DistanciaKm = &d
		}
		out = append(out, dto)
	}

	if hayUbicacion {
		// Las que no tienen coordenadas se relegan al final.
		sort.SliceStable(out, func(i, j int) bool {
			ai, aj := out[i].DistanciaKm, out[j].DistanciaKm
			if ai != nil && aj != nil {
				return *ai < *aj
			}
			if ai != nil && aj == nil {
				return true
			}
			if ai == nil && aj != nil {
				return false
			}
			return out[i].Nombre < out[j].Nombre
		})
	}

	if limite > 0 && limite < len(out) {
		out = out[:limite]
	}
	return out, nil
}

// Detalle devuelve la ficha completa de una estación. Cachea por slug.
func (s *NieveService) Detalle(slug string) (*infonieve.EstacionDetalle, error) {
	clave := "detalle:" + slug
	if v := s.getCache(clave); v != nil {
		if d, ok := v.(*infonieve.EstacionDetalle); ok {
			return d, nil
		}
	}
	d, err := s.cli.Estacion(slug)
	if err != nil {
		return nil, err
	}
	s.setCache(clave, d, s.detalleTTL)
	return d, nil
}

// Regiones expone la lista de regiones soportadas.
func (s *NieveService) Regiones() []infonieve.Region {
	return infonieve.Regiones
}

// VaciarCache limpia toda la cache; útil para depuración.
func (s *NieveService) VaciarCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = make(map[string]cacheEntry)
}

// listadoCacheado devuelve el listado del scraper cacheando 10 min.
func (s *NieveService) listadoCacheado() ([]infonieve.Estacion, error) {
	const clave = "listado"
	if v := s.getCache(clave); v != nil {
		if l, ok := v.([]infonieve.Estacion); ok {
			return l, nil
		}
	}
	// Preferimos /parte-de-nieve/ porque tiene 7 columnas (incluye remontes)
	// y es más rico que /estaciones-esqui/.
	lista, err := s.cli.ParteNieve("")
	if err != nil {
		// Fallback al listado general si el parte falla.
		lista2, err2 := s.cli.ListarEstaciones()
		if err2 != nil {
			return nil, err
		}
		lista = lista2
	}
	s.setCache(clave, lista, s.listaTTL)
	return lista, nil
}

func (s *NieveService) getCache(k string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if e, ok := s.cache[k]; ok && time.Now().Before(e.expira) {
		return e.dato
	}
	return nil
}

func (s *NieveService) setCache(k string, v any, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[k] = cacheEntry{dato: v, expira: time.Now().Add(ttl)}
}

// ─── Geometría ───────────────────────────────────────────────────────────────

// ErrCoordenadasInvalidas indica que lat/lng no están dentro de los
// rangos físicos. Se devuelve a quien llame; el handler lo convierte
// en HTTP 400.
var ErrCoordenadasInvalidas = errors.New("coordenadas fuera de rango")

// CoordsValidas comprueba latitud (-90..90) y longitud (-180..180).
func CoordsValidas(lat, lng float64) bool {
	if math.IsNaN(lat) || math.IsNaN(lng) {
		return false
	}
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}

// DistanciaHaversineKm calcula la distancia en línea recta sobre la
// superficie de la Tierra entre dos puntos en km. La fórmula de
// Haversine es precisa (<0.5 % error) para distancias terrestres y
// no requiere proyecciones cartográficas.
func DistanciaHaversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const radioTierraKm = 6371.0
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLng := (lng2 - lng1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return radioTierraKm * c
}
