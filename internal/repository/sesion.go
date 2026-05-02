// Package repository — tabla `sessions` (cookies HttpOnly persistidas).
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"skihub/internal/models"
)

// SesionRepo encapsula la persistencia de sesiones.
type SesionRepo struct {
	BD *sql.DB
}

// NuevoSesionRepo construye el repositorio.
func NuevoSesionRepo(bd *sql.DB) *SesionRepo {
	return &SesionRepo{BD: bd}
}

// Crear persiste una nueva sesión.
func (r *SesionRepo) Crear(ctx context.Context, s *models.Sesion) error {
	_, err := r.BD.ExecContext(ctx,
		`INSERT INTO sessions (token, user_id, created_at, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		s.Token, s.UsuarioID, s.CreadaEn, s.ExpiraEn,
	)
	if err != nil {
		return fmt.Errorf("insert sesion: %w", err)
	}
	return nil
}

// BuscarPorToken devuelve la sesión, o sql.ErrNoRows si no existe.
func (r *SesionRepo) BuscarPorToken(ctx context.Context, token string) (*models.Sesion, error) {
	s := &models.Sesion{}
	err := r.BD.QueryRowContext(ctx,
		`SELECT token, user_id, created_at, expires_at
		 FROM sessions WHERE token = $1`, token,
	).Scan(&s.Token, &s.UsuarioID, &s.CreadaEn, &s.ExpiraEn)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Eliminar borra una sesión por su token (logout).
func (r *SesionRepo) Eliminar(ctx context.Context, token string) error {
	_, err := r.BD.ExecContext(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}

// LimpiarExpiradas elimina las sesiones caducadas (mantenimiento).
func (r *SesionRepo) LimpiarExpiradas(ctx context.Context) error {
	_, err := r.BD.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < NOW()`)
	return err
}
