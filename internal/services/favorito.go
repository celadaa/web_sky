package services

import (
	"context"

	"skihub/internal/models"
	"skihub/internal/repository"
)

// FavoritoService implementa la lógica de negocio de favoritos.
type FavoritoService struct {
	Repo *repository.FavoritoRepo
}

// NuevoFavoritoService construye el servicio.
func NuevoFavoritoService(repo *repository.FavoritoRepo) *FavoritoService {
	return &FavoritoService{Repo: repo}
}

// Toggle invierte el estado: si no era favorita la añade, y viceversa.
// Devuelve el estado final (true = queda como favorita).
func (s *FavoritoService) Toggle(ctx context.Context, usuarioID, estacionID int64) (bool, error) {
	existe, err := s.Repo.Existe(ctx, usuarioID, estacionID)
	if err != nil {
		return false, err
	}
	if existe {
		if err := s.Repo.Quitar(ctx, usuarioID, estacionID); err != nil {
			return false, err
		}
		return false, nil
	}
	if err := s.Repo.Agregar(ctx, usuarioID, estacionID); err != nil {
		return false, err
	}
	return true, nil
}

// ListarDeUsuario devuelve las estaciones marcadas como favoritas.
func (s *FavoritoService) ListarDeUsuario(ctx context.Context, usuarioID int64) ([]models.Estacion, error) {
	return s.Repo.ListarDeUsuario(ctx, usuarioID)
}

// IDsDeUsuario devuelve los IDs marcados como favoritos.
func (s *FavoritoService) IDsDeUsuario(ctx context.Context, usuarioID int64) (map[int64]bool, error) {
	return s.Repo.IDsDeUsuario(ctx, usuarioID)
}

// ContarPorUsuario devuelve el nº de favoritos por usuario.
func (s *FavoritoService) ContarPorUsuario(ctx context.Context) (map[int64]int, error) {
	return s.Repo.ContarPorUsuario(ctx)
}
