// Package repository — acceso a la tabla `news` de PostgreSQL.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"skihub/internal/models"
)

// NoticiaRepo lee noticias de PostgreSQL.
type NoticiaRepo struct {
	BD *sql.DB
}

// NuevoNoticiaRepo construye el repositorio.
func NuevoNoticiaRepo(bd *sql.DB) *NoticiaRepo {
	return &NoticiaRepo{BD: bd}
}

// ListarRecientes devuelve las noticias ordenadas por fecha descendente.
//
// Mapeo de columnas:
//
//	id              ↔ ID
//	title           ↔ Titulo
//	summary         ↔ Extracto
//	category        ↔ Categoria
//	category_class  ↔ CategoriaClase
//	published_at    ↔ Fecha
//	image_url       ↔ Imagen
func (r *NoticiaRepo) ListarRecientes(ctx context.Context) ([]models.Noticia, error) {
	rows, err := r.BD.QueryContext(ctx, `
		SELECT id, title, summary, category, category_class, published_at, image_url
		FROM news
		ORDER BY published_at DESC, id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query news: %w", err)
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
