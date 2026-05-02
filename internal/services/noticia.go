package services

import (
	"context"

	"skihub/internal/models"
	"skihub/internal/repository"
)

// NoticiaService gestiona el listado de noticias.
type NoticiaService struct {
	Repo *repository.NoticiaRepo
}

// NuevoNoticiaService construye el servicio.
func NuevoNoticiaService(repo *repository.NoticiaRepo) *NoticiaService {
	return &NoticiaService{Repo: repo}
}

// Listar devuelve todas las noticias, las más recientes primero.
func (s *NoticiaService) Listar(ctx context.Context) ([]models.Noticia, error) {
	return s.Repo.ListarRecientes(ctx)
}
