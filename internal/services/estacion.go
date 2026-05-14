package services

import (
	"context"
	"hash/fnv"
	"strconv"
	"time"

	"skihub/internal/models"
	"skihub/internal/repository"
)

// EstacionService centraliza la lógica relacionada con estaciones.
type EstacionService struct {
	Repo    *repository.EstacionRepo
	FavRepo *repository.FavoritoRepo // opcional, para marcar favoritas
}

// NuevoEstacionService construye el servicio. Si favRepo es nil, no se
// marcan favoritas.
func NuevoEstacionService(repo *repository.EstacionRepo, favRepo *repository.FavoritoRepo) *EstacionService {
	return &EstacionService{Repo: repo, FavRepo: favRepo}
}

// Listar devuelve todas las estaciones ordenadas por distancia. Si
// usuarioID > 0 marca las que son favoritas para ese usuario.
func (s *EstacionService) Listar(ctx context.Context, usuarioID int64) ([]models.Estacion, error) {
	lista, err := s.Repo.ListarPorDistancia(ctx)
	if err != nil {
		return nil, err
	}
	if usuarioID > 0 && s.FavRepo != nil {
		ids, err := s.FavRepo.IDsDeUsuario(ctx, usuarioID)
		if err != nil {
			return nil, err
		}
		for i := range lista {
			if ids[lista[i].ID] {
				lista[i].EsFavorita = true
			}
		}
	}
	for i := range lista {
		enriquecerParte(&lista[i])
	}
	return lista, nil
}

// Obtener devuelve la ficha de una estación, marcándola como favorita
// si procede.
func (s *EstacionService) Obtener(ctx context.Context, id, usuarioID int64) (*models.Estacion, error) {
	e, err := s.Repo.ObtenerPorID(ctx, id)
	if err != nil {
		return nil, err
	}
	if usuarioID > 0 && s.FavRepo != nil {
		fav, err := s.FavRepo.Existe(ctx, usuarioID, id)
		if err != nil {
			return nil, err
		}
		e.EsFavorita = fav
	}
	enriquecerParte(e)
	return e, nil
}

// ResumenHome calcula la estación más cercana, la más lejana y la
// distancia promedio para el panel de estadísticas de la home.
func (s *EstacionService) ResumenHome(ctx context.Context) (cercana, lejana *models.Estacion, promedio float64, total int, err error) {
	lista, err := s.Repo.ListarPorDistancia(ctx)
	if err != nil || len(lista) == 0 {
		return nil, nil, 0, 0, err
	}
	var suma float64
	for i := range lista {
		suma += lista[i].Distancia
	}
	c := lista[0]
	l := lista[len(lista)-1]
	enriquecerParte(&c)
	enriquecerParte(&l)
	return &c, &l, suma / float64(len(lista)), len(lista), nil
}

// enriquecerParte rellena los campos derivados del "parte de nieve"
// (NieveMin/Max, Viento, ParteActualizado). No tocan la BD: son datos
// orientativos para la demo académica, generados de forma estable a
// partir del ID para que las cifras se mantengan constantes entre
// peticiones del mismo proceso.
func enriquecerParte(e *models.Estacion) {
	if e == nil {
		return
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte{
		byte(e.ID), byte(e.ID >> 8), byte(e.ID >> 16), byte(e.ID >> 24),
	})
	semilla := h.Sum32()

	// Rango de nieve: alrededor del valor base.
	base := e.NieveBase
	if base < 0 {
		base = 0
	}
	jitterMin := int(semilla%21) - 10 // -10..+10
	jitterMax := int((semilla>>5)%26) + 10
	min := base - 15 + jitterMin
	if min < 0 {
		min = 0
	}
	max := base + jitterMax
	if max < min+5 {
		max = min + 5
	}
	e.NieveMin = min
	e.NieveMax = max

	// Viento: 5..40 km/h estable según hash.
	viento := 5 + int((semilla>>11)%36)
	e.Viento = formatearViento(viento)

	// Parte actualizado: ahora redondeado a 15 minutos hacia abajo,
	// menos un pequeño desfase determinista (0..14 min) para que cada
	// estación muestre minutos distintos.
	desfase := time.Duration((semilla>>17)%15) * time.Minute
	t := time.Now().Add(-desfase)
	t = t.Truncate(15 * time.Minute)
	e.ParteActualizado = t
}

func formatearViento(kmh int) string {
	// pequeña narrativa descriptiva además del número
	switch {
	case kmh < 10:
		return formatNumeroKmH(kmh) + " (calmo)"
	case kmh < 20:
		return formatNumeroKmH(kmh)
	case kmh < 30:
		return formatNumeroKmH(kmh) + " (moderado)"
	default:
		return formatNumeroKmH(kmh) + " (fuerte)"
	}
}

func formatNumeroKmH(kmh int) string {
	return strconv.Itoa(kmh) + " km/h"
}
