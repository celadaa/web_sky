package services

import (
	"skihub/internal/models"
	"skihub/internal/repository"
)

// FavoritoService implementa la lógica de negocio de favoritos.
type FavoritoService struct {
	Repo *repository.FavoritoRepo
}

func NuevoFavoritoService(repo *repository.FavoritoRepo) *FavoritoService {
	return &FavoritoService{Repo: repo}
}

// Toggle invierte el estado de favorita: si no lo era la añade, y viceversa.
// Devuelve el estado final (true = queda como favorita).
func (s *FavoritoService) Toggle(usuarioID, estacionID int64) (bool, error) {
	existe, err := s.Repo.Existe(usuarioID, estacionID)
	if err != nil {
		return false, err
	}
	if existe {
		if err := s.Repo.Quitar(usuarioID, estacionID); err != nil {
			return false, err
		}
		return false, nil
	}
	if err := s.Repo.Agregar(usuarioID, estacionID); err != nil {
		return false, err
	}
	return true, nil
}

// ListarDeUsuario devuelve las estaciones marcadas como favoritas por el usuario.
func (s *FavoritoService) ListarDeUsuario(usuarioID int64) ([]models.Estacion, error) {
	return s.Repo.ListarDeUsuario(usuarioID)
}

// IDsDeUsuario devuelve un conjunto con los IDs de las favoritas del usuario,
// útil para marcar el corazón lleno/vacío en los listados.
func (s *FavoritoService) IDsDeUsuario(usuarioID int64) (map[int64]bool, error) {
	return s.Repo.IDsDeUsuario(usuarioID)
}

// ContarPorUsuario devuelve un mapa con el número de favoritos por usuario.
func (s *FavoritoService) ContarPorUsuario() (map[int64]int, error) {
	return s.Repo.ContarPorUsuario()
}
