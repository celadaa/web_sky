package services

import (
	"skihub/internal/models"
	"skihub/internal/repository"
)

// NoticiaService gestiona el listado de noticias.
type NoticiaService struct {
	Repo *repository.NoticiaRepo
}

func NuevoNoticiaService(repo *repository.NoticiaRepo) *NoticiaService {
	return &NoticiaService{Repo: repo}
}

// Listar devuelve todas las noticias, más recientes primero.
func (s *NoticiaService) Listar() ([]models.Noticia, error) {
	return s.Repo.ListarRecientes()
}
