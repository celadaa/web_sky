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

// Listar devuelve todas las noticias, las mas recientes primero.
func (s *NoticiaService) Listar(ctx context.Context) ([]models.Noticia, error) {
	return s.Repo.ListarRecientes(ctx)
}

// ListarConFiltro devuelve noticias filtradas por category_class.
// Si categoria esta vacia o es "todas", devuelve todas.
func (s *NoticiaService) ListarConFiltro(ctx context.Context, categoria string) ([]models.Noticia, error) {
	return s.Repo.ListarConFiltro(ctx, categoria)
}
