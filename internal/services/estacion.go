package services

import (
	"context"

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
	return &c, &l, suma / float64(len(lista)), len(lista), nil
}
