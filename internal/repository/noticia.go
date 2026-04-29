package repository

import (
	"database/sql"
	"fmt"

	"skihub/internal/models"
)

// NoticiaRepo lee noticias de la BD.
type NoticiaRepo struct {
	BD *sql.DB
}

func NuevoNoticiaRepo(bd *sql.DB) *NoticiaRepo {
	return &NoticiaRepo{BD: bd}
}

// ListarRecientes devuelve las noticias ordenadas por fecha descendente.
func (r *NoticiaRepo) ListarRecientes() ([]models.Noticia, error) {
	rows, err := r.BD.Query(`
		SELECT id, titulo, extracto, categoria, categoria_clase, fecha, imagen
		FROM noticias
		ORDER BY fecha DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query noticias: %w", err)
	}
	defer rows.Close()

	var lista []models.Noticia
	for rows.Next() {
		var n models.Noticia
		if err := rows.Scan(&n.ID, &n.Titulo, &n.Extracto, &n.Categoria,
			&n.CategoriaClase, &n.Fecha, &n.Imagen); err != nil {
			return nil, err
		}
		lista = append(lista, n)
	}
	return lista, rows.Err()
}
