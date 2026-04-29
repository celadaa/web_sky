package repository

import (
	"database/sql"
	"fmt"

	"skihub/internal/models"
)

// SesionRepo maneja la tabla `sesiones`.
type SesionRepo struct {
	BD *sql.DB
}

func NuevoSesionRepo(bd *sql.DB) *SesionRepo {
	return &SesionRepo{BD: bd}
}

// Crear persiste una nueva sesión.
func (r *SesionRepo) Crear(s *models.Sesion) error {
	_, err := r.BD.Exec(
		`INSERT INTO sesiones (token, usuario_id, creada_en, expira_en)
		 VALUES (?, ?, ?, ?)`,
		s.Token, s.UsuarioID, s.CreadaEn, s.ExpiraEn,
	)
	if err != nil {
		return fmt.Errorf("insert sesion: %w", err)
	}
	return nil
}

// BuscarPorToken devuelve la sesión con ese token, o sql.ErrNoRows si no existe.
func (r *SesionRepo) BuscarPorToken(token string) (*models.Sesion, error) {
	s := &models.Sesion{}
	err := r.BD.QueryRow(
		`SELECT token, usuario_id, creada_en, expira_en
		 FROM sesiones WHERE token = ?`, token,
	).Scan(&s.Token, &s.UsuarioID, &s.CreadaEn, &s.ExpiraEn)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Eliminar borra la sesión identificada por su token (logout).
func (r *SesionRepo) Eliminar(token string) error {
	_, err := r.BD.Exec(`DELETE FROM sesiones WHERE token = ?`, token)
	return err
}

// LimpiarExpiradas elimina todas las sesiones caducadas (tarea de mantenimiento).
func (r *SesionRepo) LimpiarExpiradas() error {
	_, err := r.BD.Exec(`DELETE FROM sesiones WHERE expira_en < CURRENT_TIMESTAMP`)
	return err
}
