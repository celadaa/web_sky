// Package services — generación, hash y comprobación del token de
// verificación de email.
//
// Reglas:
//   - El token "en claro" se genera con crypto/rand (32 bytes) y se
//     codifica en base64 URL-safe (sin padding) — 43 caracteres.
//   - En BD guardamos sólo el SHA-256 hex del token. Nunca el token en
//     claro: si la base se filtrara, los enlaces ya enviados no serían
//     utilizables.
//   - Caducidad: 24 horas (constante VerificacionTTL).
//   - El SHA-256 es seguro contra colisiones para este caso de uso
//     (sólo necesitamos preimagen-resistencia y los tokens son aleatorios).
package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"skihub/internal/models"
)

// VerificacionTTL es la ventana de validez del token de verificación
// de email (24h por defecto, según spec).
const VerificacionTTL = 24 * time.Hour

// Tamaño del token aleatorio antes de codificar (32 bytes = 256 bits).
const tokenBytes = 32

// ErrTokenInvalido se devuelve cuando el token no existe en BD o expiró.
// El handler debe traducirlo a un mensaje genérico para no revelar cuál
// de las dos cosas falló.
var ErrTokenInvalido = errors.New("token de verificación inválido o caducado")

// GenerarTokenVerificacion produce un par (plano, hash) seguro.
//
// El plano se mete en la URL del correo. El hash se almacena en BD. El
// handler de /confirmar-email vuelve a hashear el token entrante y busca
// el hash en BD — nunca compara el plano.
func GenerarTokenVerificacion() (plano, hash string, err error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("rand.Read: %w", err)
	}
	plano = base64.RawURLEncoding.EncodeToString(b)
	hash = HashToken(plano)
	return plano, hash, nil
}

// HashToken devuelve el SHA-256 hex de un token. Función pura — útil
// también para tests.
func HashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}

// ConfirmarEmail intenta marcar el email del usuario como verificado a
// partir de un token enviado por correo.
//
// Pasos:
//  1. Hashea el token plano.
//  2. Busca el usuario cuyo verification_token_hash == hash Y cuyo
//     verification_token_expires_at > NOW().
//  3. Marca email_verified=TRUE, limpia el token.
//  4. Devuelve el usuario actualizado (sin el token).
//
// Si no hay coincidencia o el token caducó, devuelve ErrTokenInvalido.
func (s *UsuarioService) ConfirmarEmail(ctx context.Context, tokenPlano string) (*models.Usuario, error) {
	if tokenPlano == "" {
		return nil, ErrTokenInvalido
	}
	hash := HashToken(tokenPlano)
	u, err := s.Repo.BuscarPorTokenVerificacion(ctx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenInvalido
		}
		return nil, fmt.Errorf("buscar token: %w", err)
	}
	if err := s.Repo.MarcarEmailVerificado(ctx, u.ID); err != nil {
		return nil, fmt.Errorf("marcar verificado: %w", err)
	}
	// Devolvemos los datos actualizados (sin volver a hacer una query
	// completa: ajustamos los campos en memoria).
	now := time.Now().UTC()
	u.EmailVerificado = true
	u.EmailVerificadoEn = &now
	u.VerificationTokenHash = nil
	u.VerificationTokenExpiresAt = nil
	return u, nil
}

// RegistrarConVerificacion combina:
//  1. Validar y crear el usuario (mismo flujo que Registrar).
//  2. Generar un token + hash y guardar el hash en BD.
//  3. Devolver al usuario + el token EN CLARO (para que el handler lo
//     pase al EmailService).
//
// El token NUNCA se persiste en claro. El handler debe garantizar que
// no se loguea.
func (s *UsuarioService) RegistrarConVerificacion(ctx context.Context, d DatosRegistro) (*models.Usuario, string, error) {
	u, err := s.Registrar(ctx, d)
	if err != nil {
		return nil, "", err
	}
	plano, hash, err := GenerarTokenVerificacion()
	if err != nil {
		return nil, "", fmt.Errorf("token: %w", err)
	}
	expira := time.Now().Add(VerificacionTTL)
	if err := s.Repo.GuardarTokenVerificacion(ctx, u.ID, hash, expira); err != nil {
		return nil, "", fmt.Errorf("guardar token: %w", err)
	}
	u.VerificationTokenHash = &hash
	u.VerificationTokenExpiresAt = &expira
	return u, plano, nil
}

// ReenviarVerificacion genera un nuevo token para un usuario ya creado
// (típicamente cuando el primer correo no llegó o expiró). El token
// antiguo queda invalidado al sobrescribirse.
//
// El handler debe haber comprobado antes que el usuario existe y que
// NO está ya verificado.
func (s *UsuarioService) ReenviarVerificacion(ctx context.Context, usuarioID int64) (string, error) {
	plano, hash, err := GenerarTokenVerificacion()
	if err != nil {
		return "", err
	}
	expira := time.Now().Add(VerificacionTTL)
	if err := s.Repo.GuardarTokenVerificacion(ctx, usuarioID, hash, expira); err != nil {
		return "", err
	}
	return plano, nil
}
