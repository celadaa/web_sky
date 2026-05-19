// Package repository — acceso a las tablas `lodgings` y `lodging_bookings`.
//
// Mapeo de columnas SQL (inglés) ↔ campos del modelo (español):
//
//   lodgings.id              ↔ ID
//   lodgings.station_id      ↔ EstacionID
//   lodgings.name            ↔ Nombre
//   lodgings.kind            ↔ Tipo
//   lodgings.image_url       ↔ Imagen
//   lodgings.distance_km     ↔ Distancia
//   lodgings.rating          ↔ Valoracion
//   lodgings.price_eur       ↔ PrecioNoche
//   lodgings.zone            ↔ Zona
//   lodgings.description     ↔ Descripcion
//
// Si la migración 012 no se ha aplicado todavía, ListarPorEstacion
// devuelve sql.ErrNoRows en la primera consulta (que el servicio
// envuelve a una lista vacía) — la app sigue funcionando, la sección
// de alojamientos simplemente no aparece.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"skihub/internal/models"
)

// AlojamientoRepo lee/escribe alojamientos y reservas en PostgreSQL.
type AlojamientoRepo struct {
	BD *sql.DB
}

// NuevoAlojamientoRepo construye el repositorio.
func NuevoAlojamientoRepo(bd *sql.DB) *AlojamientoRepo {
	return &AlojamientoRepo{BD: bd}
}

const columnasAlojamiento = `
	id, station_id, name, kind, image_url, distance_km,
	rating, price_eur, zone, description, created_at, updated_at
`

func scanAlojamiento(scanner interface {
	Scan(...any) error
}, a *models.Alojamiento) error {
	return scanner.Scan(
		&a.ID, &a.EstacionID, &a.Nombre, &a.Tipo, &a.Imagen,
		&a.Distancia, &a.Valoracion, &a.PrecioNoche, &a.Zona,
		&a.Descripcion, &a.CreadoEn, &a.ActualizadoEn,
	)
}

// ListarPorEstacion devuelve los alojamientos asociados a una estación,
// ordenados por distancia ascendente.
func (r *AlojamientoRepo) ListarPorEstacion(ctx context.Context, estacionID int64) ([]models.Alojamiento, error) {
	rows, err := r.BD.QueryContext(ctx, `
		SELECT `+columnasAlojamiento+`
		FROM lodgings
		WHERE station_id = $1
		ORDER BY distance_km ASC, price_eur ASC
	`, estacionID)
	if err != nil {
		return nil, fmt.Errorf("query lodgings: %w", err)
	}
	defer rows.Close()

	var out []models.Alojamiento
	for rows.Next() {
		var a models.Alojamiento
		if err := scanAlojamiento(rows, &a); err != nil {
			return nil, fmt.Errorf("scan lodging: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ObtenerPorID devuelve un alojamiento por su PK. También trae el nombre
// de la estación asociada (JOIN) para mostrarlo en la página de detalle
// sin tener que hacer una segunda consulta.
func (r *AlojamientoRepo) ObtenerPorID(ctx context.Context, id int64) (*models.Alojamiento, error) {
	row := r.BD.QueryRowContext(ctx, `
		SELECT `+columnasAlojamiento+`, s.name
		FROM lodgings l
		JOIN stations s ON s.id = l.station_id
		WHERE l.id = $1
	`, id)
	var a models.Alojamiento
	err := row.Scan(
		&a.ID, &a.EstacionID, &a.Nombre, &a.Tipo, &a.Imagen,
		&a.Distancia, &a.Valoracion, &a.PrecioNoche, &a.Zona,
		&a.Descripcion, &a.CreadoEn, &a.ActualizadoEn,
		&a.NombreEstacion,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// InsertarReserva persiste una reserva nueva y devuelve su ID.
// El servicio se encarga de calcular noches y total antes de llamar
// aquí; esta función no valida nada (solo INSERT).
func (r *AlojamientoRepo) InsertarReserva(ctx context.Context, res *models.ReservaAlojamiento) (int64, error) {
	row := r.BD.QueryRowContext(ctx, `
		INSERT INTO lodging_bookings (
			user_id, lodging_id, lodging_name, lodging_kind,
			station_id, station_name,
			check_in, check_out, nights, guests,
			price_per_night, total_eur, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		RETURNING id
	`,
		res.UsuarioID, res.AlojamientoID, res.NombreAloja, res.TipoAloja,
		res.EstacionID, res.NombreEstacion,
		res.FechaEntrada, res.FechaSalida, res.Noches, res.Huespedes,
		res.PrecioNoche, res.TotalEur, res.Estado,
	)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("insert lodging_booking: %w", err)
	}
	return id, nil
}
