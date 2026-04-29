package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"skihub/internal/models"
	"skihub/internal/repository"
)

// Duración por defecto de una sesión: 7 días.
const DuracionSesion = 7 * 24 * time.Hour

// SesionService gestiona el ciclo de vida de las sesiones.
type SesionService struct {
	SesRepo *repository.SesionRepo
	UsrRepo *repository.UsuarioRepo
}

// NuevoSesionService construye el servicio.
func NuevoSesionService(ses *repository.SesionRepo, usr *repository.UsuarioRepo) *SesionService {
	return &SesionService{SesRepo: ses, UsrRepo: usr}
}

// Crear genera una nueva sesión para el usuario indicado y devuelve el token.
func (s *SesionService) Crear(usuarioID int64) (*models.Sesion, error) {
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
	if err := s.SesRepo.Crear(sesion); err != nil {
		return nil, err
	}
	return sesion, nil
}

// UsuarioDeToken devuelve el usuario dueño de la sesión identificada por token,
// o nil si la sesión no existe o está expirada.
func (s *SesionService) UsuarioDeToken(token string) (*models.Usuario, error) {
	if token == "" {
		return nil, nil
	}
	sesion, err := s.SesRepo.BuscarPorToken(token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if !sesion.Vigente() {
		// Limpia al paso las sesiones muertas.
		_ = s.SesRepo.Eliminar(token)
		return nil, nil
	}
	// Recupera el usuario por su ID a través del repo de usuarios.
	// (Añadimos un helper directo ObtenerPorID para esto.)
	u, err := obtenerPorID(s.UsrRepo, sesion.UsuarioID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Cerrar invalida la sesión (logout).
func (s *SesionService) Cerrar(token string) error {
	if token == "" {
		return nil
	}
	return s.SesRepo.Eliminar(token)
}

// generarToken produce un token seguro de 32 bytes en hexadecimal (64 caracteres).
func generarToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// obtenerPorID es un wrapper que evita exponer un método nuevo si no se desea,
// pero en la práctica se añade como método del repo. Aquí lo dejamos como
// función privada que pregunta directamente al *UsuarioRepo.
func obtenerPorID(r *repository.UsuarioRepo, id int64) (*models.Usuario, error) {
	return r.ObtenerPorID(id)
}
