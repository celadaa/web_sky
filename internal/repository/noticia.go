// Package repository — acceso a la tabla news de PostgreSQL.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"skihub/internal/models"
)

// NoticiaRepo lee y escribe noticias en PostgreSQL.
type NoticiaRepo struct {
	BD *sql.DB
}

// NuevoNoticiaRepo construye el repositorio.
func NuevoNoticiaRepo(bd *sql.DB) *NoticiaRepo {
	return &NoticiaRepo{BD: bd}
}

// ListarRecientes devuelve todas las noticias ordenadas por fecha descendente.
func (r *NoticiaRepo) ListarRecientes(ctx context.Context) ([]models.Noticia, error) {
	return r.ListarConFiltro(ctx, "")
}

// ListarConFiltro devuelve noticias filtradas por category_class.
// categoria="" o "todas" devuelve todas. Maximo 60 resultados.
func (r *NoticiaRepo) ListarConFiltro(ctx context.Context, categoria string) ([]models.Noticia, error) {
	const base = `
		SELECT id, title, summary, category, category_class, published_at,
		       image_url, source_name, source_url, original_url, slug, is_external
		FROM news`

	var (
		rows *sql.Rows
		err  error
	)

	if categoria != "" && categoria != "todas" {
		rows, err = r.BD.QueryContext(ctx, base+`
			WHERE category_class = $1
			ORDER BY published_at DESC, id DESC
			LIMIT 60`, categoria)
	} else {
		rows, err = r.BD.QueryContext(ctx, base+`
			ORDER BY published_at DESC, id DESC
			LIMIT 60`)
	}
	if err != nil {
		return nil, fmt.Errorf("query news: %w", err)
	}
	defer rows.Close()

	var lista []models.Noticia
	for rows.Next() {
		var n models.Noticia
		if err := rows.Scan(
			&n.ID, &n.Titulo, &n.Extracto, &n.Categoria,
			&n.CategoriaClase, &n.Fecha, &n.Imagen,
			&n.SourceName, &n.SourceURL, &n.OriginalURL,
			&n.Slug, &n.IsExternal,
		); err != nil {
			return nil, fmt.Errorf("scan noticia: %w", err)
		}
		lista = append(lista, n)
	}
	return lista, rows.Err()
}

// UpsertExterno inserta una noticia externa o actualiza titulo, extracto e
// imagen si ya existe (detectada por original_url).
// Devuelve (true, nil) si se inserto una fila nueva.
// Devuelve (false, nil) si ya existia y se actualizo.
func (r *NoticiaRepo) UpsertExterno(ctx context.Context, n models.Noticia) (nuevo bool, err error) {
	const q = `
		INSERT INTO news
			(title, summary, category, category_class, published_at,
			 image_url, source_name, source_url, original_url, slug, is_external)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true)
		ON CONFLICT (original_url) WHERE original_url <> ''
		DO UPDATE SET
			title          = EXCLUDED.title,
			summary        = EXCLUDED.summary,
			image_url      = CASE
				                WHEN news.image_url = '' THEN EXCLUDED.image_url
				                ELSE news.image_url
			                END,
			published_at   = EXCLUDED.published_at,
			updated_at     = NOW()
		RETURNING (xmax = 0) AS inserted`

	var inserted bool
	err = r.BD.QueryRowContext(ctx, q,
		n.Titulo, n.Extracto, n.Categoria, n.CategoriaClase, n.Fecha,
		n.Imagen, n.SourceName, n.SourceURL, n.OriginalURL, n.Slug,
	).Scan(&inserted)
	if err != nil {
		return false, fmt.Errorf("upsert noticia externa %q: %w", n.OriginalURL, err)
	}
	return inserted, nil
}
