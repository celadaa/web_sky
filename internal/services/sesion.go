package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"skihub/internal/models"
	"skihub/internal/repository"
)

// DuracionSesion es la vida útil por defecto de una sesión: 7 días.
const DuracionSesion = 7 * 24 * time.Hour

// SesionService gestiona el ciclo de vida de las sesiones HTTP.
type SesionService struct {
	SesRepo *repository.SesionRepo
	UsrRepo *repository.UsuarioRepo
}

// NuevoSesionService construye el servicio.
func NuevoSesionService(ses *repository.SesionRepo, usr *repository.UsuarioRepo) *SesionService {
	return &SesionService{SesRepo: ses, UsrRepo: usr}
}

// Crear genera una nueva sesión y la persiste.
func (s *SesionService) Crear(ctx context.Context, usuarioID int64) (*models.Sesion, error) {
	token, err := generarToken()
	if err != nil {
		return nil, fmt.Errorf("generando token: %w", err)
	}
	ahora := time.Now()
	sesion := &models.Sesion{
		Token:     token,
		UsuarioID: usuarioID,
		CreadaEn:  ahora,
		ExpiraEn:  ahora.Add(DuracionSesion),
	}
	if err := s.SesRepo.Crear(ctx, sesion); err != nil {
		return nil, err
	}
	return sesion, nil
}

// UsuarioDeToken devuelve el usuario asociado al token, o (nil,nil) si
// el token está vacío, no existe o ha expirado.
func (s *SesionService) UsuarioDeToken(ctx context.Context, token string) (*models.Usuario, error) {
	if token == "" {
		return nil, nil
	}
	sesion, err := s.SesRepo.BuscarPorToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if !sesion.Vigente() {
		// Limpia al paso las sesiones muertas. Ignoramos el error de
		// borrado porque no afecta al flujo del usuario.
		_ = s.SesRepo.Eliminar(ctx, token)
		return nil, nil
	}
	return s.UsrRepo.ObtenerPorID(ctx, sesion.UsuarioID)
}

// Cerrar invalida la sesión (logout).
func (s *SesionService) Cerrar(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.SesRepo.Eliminar(ctx, token)
}

// generarToken produce un token seguro de 32 bytes en hex (64 chars).
func generarToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
