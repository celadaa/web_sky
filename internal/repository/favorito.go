package repository

import (
	"database/sql"
	"fmt"

	"skihub/internal/models"
)

// FavoritoRepo maneja la tabla `favoritos` (N:M usuarios-estaciones).
type FavoritoRepo struct {
	BD *sql.DB
}

func NuevoFavoritoRepo(bd *sql.DB) *FavoritoRepo {
	return &FavoritoRepo{BD: bd}
}

// Agregar inserta la relación. Si ya existe (PRIMARY KEY compuesta) no hace nada.
func (r *FavoritoRepo) Agregar(usuarioID, estacionID int64) error {
	_, err := r.BD.Exec(
		`INSERT OR IGNORE INTO favoritos (usuario_id, estacion_id) VALUES (?, ?)`,
		usuarioID, estacionID,
	)
	if err != nil {
		return fmt.Errorf("agregar favorito: %w", err)
	}
	return nil
}

// Quitar elimina la relación entre usuario y estación.
func (r *FavoritoRepo) Quitar(usuarioID, estacionID int64) error {
	_, err := r.BD.Exec(
		`DELETE FROM favoritos WHERE usuario_id = ? AND estacion_id = ?`,
		usuarioID, estacionID,
	)
	return err
}

// Existe comprueba si una estación es favorita de un usuario.
func (r *FavoritoRepo) Existe(usuarioID, estacionID int64) (bool, error) {
	var uno int
	err := r.BD.QueryRow(
		`SELECT 1 FROM favoritos WHERE usuario_id = ? AND estacion_id = ?`,
		usuarioID, estacionID,
	).Scan(&uno)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// IDsDeUsuario devuelve los IDs de las estaciones favoritas del usuario.
func (r *FavoritoRepo) IDsDeUsuario(usuarioID int64) (map[int64]bool, error) {
	rows, err := r.BD.Query(
		`SELECT estacion_id FROM favoritos WHERE usuario_id = ?`, usuarioID,
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
// Se usa en el panel de administración.
func (r *FavoritoRepo) ContarPorUsuario() (map[int64]int, error) {
	rows, err := r.BD.Query(
		`SELECT usuario_id, COUNT(*) FROM favoritos GROUP BY usuario_id`,
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

// ListarDeUsuario devuelve las estaciones favoritas completas del usuario,
// con un JOIN a la tabla estaciones, ordenadas por fecha de marcado DESC.
func (r *FavoritoRepo) ListarDeUsuario(usuarioID int64) ([]models.Estacion, error) {
	rows, err := r.BD.Query(`
		SELECT e.id, e.nombre, e.ubicacion, e.distancia, e.temperatura,
		       e.nieve_base, e.nieve_nueva, e.pistas_abiertas, e.pistas_totales,
		       e.remontes_op, e.remontes_tot, e.ultima_nevada, e.altitud,
		       e.km_esquiables, e.dificultad, e.telefono, e.imagen, e.descripcion
		FROM favoritos f
		JOIN estaciones e ON e.id = f.estacion_id
		WHERE f.usuario_id = ?
		ORDER BY f.creada_en DESC
	`, usuarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []models.Estacion
	for rows.Next() {
		var e models.Estacion
		if err := rows.Scan(&e.ID, &e.Nombre, &e.Ubicacion, &e.Distancia,
			&e.Temperatura, &e.NieveBase, &e.NieveNueva,
			&e.PistasAbiertas, &e.PistasTotales, &e.RemontesOp, &e.RemontesTot,
			&e.UltimaNevada, &e.Altitud, &e.KmEsquiables, &e.Dificultad,
			&e.Telefono, &e.Imagen, &e.Descripcion); err != nil {
			return nil, err
		}
		e.EsFavorita = true
		lista = append(lista, e)
	}
	return lista, rows.Err()
}
