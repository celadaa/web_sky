// Package repository — acceso a las tablas `hotels` y `hotel_images`.
//
// Mapeo de columnas SQL (inglés) ↔ campos del modelo (español):
//
//	id                     ↔ ID
//	station_id             ↔ EstacionID
//	name                   ↔ Nombre
//	slug                   ↔ Slug
//	description            ↔ Descripcion
//	address                ↔ Direccion
//	city                   ↔ Ciudad
//	country                ↔ Pais
//	distance_to_station_km ↔ DistanciaEstacion
//	price_from             ↔ PrecioDesde
//	currency               ↔ Divisa
//	rating                 ↔ Rating
//	review_count           ↔ NumResenas
//	amenities              ↔ Amenities
//	booking_url            ↔ BookingURL
//	official_url           ↔ OficialURL
//	main_image_url         ↔ ImagenURL
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"skihub/internal/models"
)

// HotelRepo lee y escribe hoteles en PostgreSQL.
type HotelRepo struct {
	BD *sql.DB
}

// NuevoHotelRepo construye el repositorio.
func NuevoHotelRepo(bd *sql.DB) *HotelRepo {
	return &HotelRepo{BD: bd}
}

// columnasHotel centraliza la lista de columnas (prefijadas con h. para
// poder usarlas en JOINs) y mantiene el orden coherente con scanHotel().
const columnasHotel = `
	h.id, h.station_id, h.name, h.slug, h.description,
	h.address, h.city, h.country, h.distance_to_station_km,
	h.price_from, h.currency, h.rating, h.review_count,
	h.amenities, h.booking_url, h.official_url, h.main_image_url,
	h.created_at, h.updated_at, s.name AS station_name
`

// scanHotel vuelca la fila actual en un models.Hotel.
func scanHotel(scanner interface {
	Scan(...any) error
}, h *models.Hotel) error {
	return scanner.Scan(
		&h.ID, &h.EstacionID, &h.Nombre, &h.Slug, &h.Descripcion,
		&h.Direccion, &h.Ciudad, &h.Pais, &h.DistanciaEstacion,
		&h.PrecioDesde, &h.Divisa, &h.Rating, &h.NumResenas,
		&h.Amenities, &h.BookingURL, &h.OficialURL, &h.ImagenURL,
		&h.CreadoEn, &h.Actualizado, &h.EstacionNombre,
	)
}

// FiltroHoteles agrupa los filtros opcionales del listado. Valor cero =
// sin filtro (0 en numéricos significa "no filtrar").
type FiltroHoteles struct {
	EstacionID   int64
	PrecioMax    float64
	DistanciaMax float64
}

// Listar devuelve los hoteles que cumplen el filtro, ordenados por
// distancia a la estación y precio.
func (r *HotelRepo) Listar(ctx context.Context, f FiltroHoteles) ([]models.Hotel, error) {
	q := `SELECT ` + columnasHotel + `
		FROM hotels h
		JOIN stations s ON s.id = h.station_id
		WHERE ($1 = 0 OR h.station_id = $1)
		  AND ($2 = 0 OR h.price_from <= $2)
		  AND ($3 = 0 OR h.distance_to_station_km <= $3)
		ORDER BY h.distance_to_station_km ASC, h.price_from ASC, h.id ASC`
	rows, err := r.BD.QueryContext(ctx, q, f.EstacionID, f.PrecioMax, f.DistanciaMax)
	if err != nil {
		return nil, fmt.Errorf("query hotels: %w", err)
	}
	defer rows.Close()

	var lista []models.Hotel
	for rows.Next() {
		var h models.Hotel
		if err := scanHotel(rows, &h); err != nil {
			return nil, err
		}
		lista = append(lista, h)
	}
	return lista, rows.Err()
}

// ObtenerPorSlug devuelve la ficha completa de un hotel por su slug.
func (r *HotelRepo) ObtenerPorSlug(ctx context.Context, slug string) (*models.Hotel, error) {
	h := &models.Hotel{}
	err := scanHotel(r.BD.QueryRowContext(ctx,
		`SELECT `+columnasHotel+` FROM hotels h JOIN stations s ON s.id = h.station_id WHERE h.slug = $1`,
		slug), h)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// ObtenerPorID devuelve la ficha completa de un hotel por id.
func (r *HotelRepo) ObtenerPorID(ctx context.Context, id int64) (*models.Hotel, error) {
	h := &models.Hotel{}
	err := scanHotel(r.BD.QueryRowContext(ctx,
		`SELECT `+columnasHotel+` FROM hotels h JOIN stations s ON s.id = h.station_id WHERE h.id = $1`,
		id), h)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// ExisteSlug indica si ya hay un hotel con ese slug (excluyendo opcionalmente
// un id, útil al editar).
func (r *HotelRepo) ExisteSlug(ctx context.Context, slug string, excluirID int64) (bool, error) {
	var n int
	err := r.BD.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hotels WHERE slug = $1 AND id <> $2`,
		slug, excluirID).Scan(&n)
	return n > 0, err
}

// Crear inserta un hotel y devuelve su id.
func (r *HotelRepo) Crear(ctx context.Context, h *models.Hotel) (int64, error) {
	var id int64
	err := r.BD.QueryRowContext(ctx, `
		INSERT INTO hotels (
			station_id, name, slug, description, address, city, country,
			distance_to_station_km, price_from, currency, rating, review_count,
			amenities, booking_url, official_url, main_image_url
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING id`,
		h.EstacionID, h.Nombre, h.Slug, h.Descripcion, h.Direccion, h.Ciudad, h.Pais,
		h.DistanciaEstacion, h.PrecioDesde, h.Divisa, h.Rating, h.NumResenas,
		h.Amenities, h.BookingURL, h.OficialURL, h.ImagenURL,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert hotel: %w", err)
	}
	return id, nil
}

// Actualizar persiste los cambios de un hotel existente.
func (r *HotelRepo) Actualizar(ctx context.Context, h *models.Hotel) error {
	_, err := r.BD.ExecContext(ctx, `
		UPDATE hotels SET
			station_id = $1, name = $2, slug = $3, description = $4,
			address = $5, city = $6, country = $7, distance_to_station_km = $8,
			price_from = $9, currency = $10, rating = $11, review_count = $12,
			amenities = $13, booking_url = $14, official_url = $15, main_image_url = $16
		WHERE id = $17`,
		h.EstacionID, h.Nombre, h.Slug, h.Descripcion,
		h.Direccion, h.Ciudad, h.Pais, h.DistanciaEstacion,
		h.PrecioDesde, h.Divisa, h.Rating, h.NumResenas,
		h.Amenities, h.BookingURL, h.OficialURL, h.ImagenURL, h.ID,
	)
	if err != nil {
		return fmt.Errorf("update hotel %d: %w", h.ID, err)
	}
	return nil
}

// Borrar elimina un hotel (sus imágenes caen por ON DELETE CASCADE).
func (r *HotelRepo) Borrar(ctx context.Context, id int64) error {
	_, err := r.BD.ExecContext(ctx, `DELETE FROM hotels WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete hotel %d: %w", id, err)
	}
	return nil
}

// ImagenesDeHotel devuelve la galería ordenada de un hotel.
func (r *HotelRepo) ImagenesDeHotel(ctx context.Context, hotelID int64) ([]models.HotelImagen, error) {
	rows, err := r.BD.QueryContext(ctx, `
		SELECT id, hotel_id, image_url, alt_text, source_name, source_url,
		       sort_order, is_primary, created_at
		FROM hotel_images
		WHERE hotel_id = $1
		ORDER BY is_primary DESC, sort_order ASC, id ASC`, hotelID)
	if err != nil {
		return nil, fmt.Errorf("query hotel_images: %w", err)
	}
	defer rows.Close()

	var lista []models.HotelImagen
	for rows.Next() {
		var img models.HotelImagen
		if err := rows.Scan(&img.ID, &img.HotelID, &img.URL, &img.AltText,
			&img.FuenteNom, &img.FuenteURL, &img.Orden, &img.EsPrimaria, &img.CreadoEn); err != nil {
			return nil, err
		}
		lista = append(lista, img)
	}
	return lista, rows.Err()
}

// ReemplazarImagenes sustituye en una transacción toda la galería del hotel
// por la lista dada (modo simple del panel admin: una URL por línea).
func (r *HotelRepo) ReemplazarImagenes(ctx context.Context, hotelID int64, imgs []models.HotelImagen) error {
	tx, err := r.BD.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM hotel_images WHERE hotel_id = $1`, hotelID); err != nil {
		return fmt.Errorf("delete hotel_images %d: %w", hotelID, err)
	}
	for i, img := range imgs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO hotel_images (hotel_id, image_url, alt_text, source_name, source_url, sort_order, is_primary)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			hotelID, img.URL, img.AltText, img.FuenteNom, img.FuenteURL, i, i == 0,
		); err != nil {
			return fmt.Errorf("insert hotel_image: %w", err)
		}
	}
	return tx.Commit()
}
