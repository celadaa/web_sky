package repository

import (
	"database/sql"
	"fmt"

	"skihub/internal/models"
)

// EstacionRepo lee estaciones de la BD.
type EstacionRepo struct {
	BD *sql.DB
}

// NuevoEstacionRepo construye el repositorio.
func NuevoEstacionRepo(bd *sql.DB) *EstacionRepo {
	return &EstacionRepo{BD: bd}
}

// ListarPorDistancia devuelve todas las estaciones ordenadas por distancia ascendente.
func (r *EstacionRepo) ListarPorDistancia() ([]models.Estacion, error) {
	rows, err := r.BD.Query(`
		SELECT id, nombre, ubicacion, distancia, temperatura, nieve_base,
		       nieve_nueva, pistas_abiertas, pistas_totales, remontes_op,
		       remontes_tot, ultima_nevada, altitud, km_esquiables,
		       dificultad, telefono, imagen, descripcion,
		       precio_adulto, precio_nino, precio_senior
		FROM estaciones
		ORDER BY distancia ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query estaciones: %w", err)
	}
	defer rows.Close()

	var lista []models.Estacion
	for rows.Next() {
		var e models.Estacion
		if err := rows.Scan(&e.ID, &e.Nombre, &e.Ubicacion, &e.Distancia,
			&e.Temperatura, &e.NieveBase, &e.NieveNueva,
			&e.PistasAbiertas, &e.PistasTotales, &e.RemontesOp, &e.RemontesTot,
			&e.UltimaNevada, &e.Altitud, &e.KmEsquiables, &e.Dificultad,
			&e.Telefono, &e.Imagen, &e.Descripcion,
			&e.PrecioAdulto, &e.PrecioNino, &e.PrecioSenior); err != nil {
			return nil, err
		}
		lista = append(lista, e)
	}
	return lista, rows.Err()
}

// ObtenerPorID devuelve la ficha completa de una estación.
func (r *EstacionRepo) ObtenerPorID(id int64) (*models.Estacion, error) {
	e := &models.Estacion{}
	err := r.BD.QueryRow(`
		SELECT id, nombre, ubicacion, distancia, temperatura, nieve_base,
		       nieve_nueva, pistas_abiertas, pistas_totales, remontes_op,
		       remontes_tot, ultima_nevada, altitud, km_esquiables,
		       dificultad, telefono, imagen, descripcion,
		       precio_adulto, precio_nino, precio_senior
		FROM estaciones WHERE id = ?`, id,
	).Scan(&e.ID, &e.Nombre, &e.Ubicacion, &e.Distancia,
		&e.Temperatura, &e.NieveBase, &e.NieveNueva,
		&e.PistasAbiertas, &e.PistasTotales, &e.RemontesOp, &e.RemontesTot,
		&e.UltimaNevada, &e.Altitud, &e.KmEsquiables, &e.Dificultad,
		&e.Telefono, &e.Imagen, &e.Descripcion,
		&e.PrecioAdulto, &e.PrecioNino, &e.PrecioSenior)
	if err != nil {
		return nil, err
	}
	return e, nil
}
