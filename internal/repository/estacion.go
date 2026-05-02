// Package repository — acceso a la tabla `stations` de PostgreSQL.
//
// Mapeo de columnas SQL (inglés) ↔ campos del modelo (español):
//
//	id              ↔ ID
//	name            ↔ Nombre
//	location        ↔ Ubicacion
//	distance_km     ↔ Distancia
//	temperature_c   ↔ Temperatura
//	snow_base_cm    ↔ NieveBase
//	snow_new_cm     ↔ NieveNueva
//	slopes_open     ↔ PistasAbiertas
//	slopes_total    ↔ PistasTotales
//	lifts_open      ↔ RemontesOp
//	lifts_total     ↔ RemontesTot
//	last_snowfall   ↔ UltimaNevada
//	altitude        ↔ Altitud
//	ski_km          ↔ KmEsquiables
//	difficulty      ↔ Dificultad
//	phone           ↔ Telefono
//	image_url       ↔ Imagen
//	description     ↔ Descripcion
//	price_adult     ↔ PrecioAdulto
//	price_child     ↔ PrecioNino
//	price_senior    ↔ PrecioSenior
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"skihub/internal/models"
)

// EstacionRepo lee estaciones de PostgreSQL.
type EstacionRepo struct {
	BD *sql.DB
}

// NuevoEstacionRepo construye el repositorio.
func NuevoEstacionRepo(bd *sql.DB) *EstacionRepo {
	return &EstacionRepo{BD: bd}
}

// columnasEstacion centraliza la lista de columnas para reutilizarla
// en las queries y mantener el orden coherente con el Scan().
const columnasEstacion = `
	id, name, location, distance_km, temperature_c, snow_base_cm,
	snow_new_cm, slopes_open, slopes_total, lifts_open, lifts_total,
	last_snowfall, altitude, ski_km, difficulty, phone, image_url,
	description, price_adult, price_child, price_senior
`

// scanEstacion vuelca la fila actual en una models.Estacion.
func scanEstacion(scanner interface {
	Scan(...any) error
}, e *models.Estacion) error {
	return scanner.Scan(
		&e.ID, &e.Nombre, &e.Ubicacion, &e.Distancia,
		&e.Temperatura, &e.NieveBase, &e.NieveNueva,
		&e.PistasAbiertas, &e.PistasTotales, &e.RemontesOp, &e.RemontesTot,
		&e.UltimaNevada, &e.Altitud, &e.KmEsquiables, &e.Dificultad,
		&e.Telefono, &e.Imagen, &e.Descripcion,
		&e.PrecioAdulto, &e.PrecioNino, &e.PrecioSenior,
	)
}

// ListarPorDistancia devuelve todas las estaciones ordenadas por distancia.
func (r *EstacionRepo) ListarPorDistancia(ctx context.Context) ([]models.Estacion, error) {
	rows, err := r.BD.QueryContext(ctx, `
		SELECT `+columnasEstacion+`
		FROM stations
		ORDER BY distance_km ASC, id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query stations: %w", err)
	}
	defer rows.Close()

	var lista []models.Estacion
	for rows.Next() {
		var e models.Estacion
		if err := scanEstacion(rows, &e); err != nil {
			return nil, err
		}
		lista = append(lista, e)
	}
	return lista, rows.Err()
}

// ObtenerPorID devuelve la ficha completa de una estación.
func (r *EstacionRepo) ObtenerPorID(ctx context.Context, id int64) (*models.Estacion, error) {
	e := &models.Estacion{}
	err := scanEstacion(
		r.BD.QueryRowContext(ctx,
			`SELECT `+columnasEstacion+` FROM stations WHERE id = $1`, id),
		e,
	)
	if err != nil {
		return nil, err
	}
	return e, nil
}
