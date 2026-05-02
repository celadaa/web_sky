// Package repository — relación N:M usuarios ↔ estaciones (tabla `favorites`).
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"skihub/internal/models"
)

// FavoritoRepo maneja la tabla `favorites`.
type FavoritoRepo struct {
	BD *sql.DB
}

// NuevoFavoritoRepo construye el repositorio.
func NuevoFavoritoRepo(bd *sql.DB) *FavoritoRepo {
	return &FavoritoRepo{BD: bd}
}

// Agregar inserta la relación. Si ya existe (PK compuesta) no falla.
func (r *FavoritoRepo) Agregar(ctx context.Context, usuarioID, estacionID int64) error {
	_, err := r.BD.ExecContext(ctx,
		`INSERT INTO favorites (user_id, station_id)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id, station_id) DO NOTHING`,
		usuarioID, estacionID,
	)
	if err != nil {
		return fmt.Errorf("agregar favorito: %w", err)
	}
	return nil
}

// Quitar elimina la relación entre usuario y estación.
func (r *FavoritoRepo) Quitar(ctx context.Context, usuarioID, estacionID int64) error {
	_, err := r.BD.ExecContext(ctx,
		`DELETE FROM favorites WHERE user_id = $1 AND station_id = $2`,
		usuarioID, estacionID,
	)
	return err
}

// Existe comprueba si esa estación es favorita del usuario.
func (r *FavoritoRepo) Existe(ctx context.Context, usuarioID, estacionID int64) (bool, error) {
	var uno int
	err := r.BD.QueryRowContext(ctx,
		`SELECT 1 FROM favorites WHERE user_id = $1 AND station_id = $2`,
		usuarioID, estacionID,
	).Scan(&uno)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// IDsDeUsuario devuelve los IDs de las estaciones favoritas del usuario.
func (r *FavoritoRepo) IDsDeUsuario(ctx context.Context, usuarioID int64) (map[int64]bool, error) {
	rows, err := r.BD.QueryContext(ctx,
		`SELECT station_id FROM favorites WHERE user_id = $1`, usuarioID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = true
	}
	return ids, rows.Err()
}

// ContarPorUsuario devuelve un mapa {usuarioID -> nº de favoritos}.
func (r *FavoritoRepo) ContarPorUsuario(ctx context.Context) (map[int64]int, error) {
	rows, err := r.BD.QueryContext(ctx,
		`SELECT user_id, COUNT(*) FROM favorites GROUP BY user_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	mapa := map[int64]int{}
	for rows.Next() {
		var id int64
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		mapa[id] = n
	}
	return mapa, rows.Err()
}

// ListarDeUsuario devuelve las estaciones favoritas del usuario con
// JOIN a stations, ordenadas por fecha de marcado descendente.
func (r *FavoritoRepo) ListarDeUsuario(ctx context.Context, usuarioID int64) ([]models.Estacion, error) {
	rows, err := r.BD.QueryContext(ctx, `
		SELECT s.id, s.name, s.location, s.distance_km, s.temperature_c,
		       s.snow_base_cm, s.snow_new_cm, s.slopes_open, s.slopes_total,
		       s.lifts_open, s.lifts_total, s.last_snowfall, s.altitude,
		       s.ski_km, s.difficulty, s.phone, s.image_url, s.description,
		       s.price_adult, s.price_child, s.price_senior
		FROM favorites f
		JOIN stations s ON s.id = f.station_id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
	`, usuarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []models.Estacion
	for rows.Next() {
		var e models.Estacion
		if err := scanEstacion(rows, &e); err != nil {
			return nil, err
		}
		e.EsFavorita = true
		lista = append(lista, e)
	}
	return lista, rows.Err()
}
