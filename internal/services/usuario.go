// Package services implementa la lógica de negocio de Snowbreak.
// Los handlers llaman a los servicios, nunca a los repositorios directamente.
package services

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"skihub/internal/models"
	"skihub/internal/repository"
)

// Errores de negocio que el handler traducirá a mensajes para el usuario.
var (
	ErrNombreInvalido   = errors.New("el nombre debe tener entre 2 y 60 caracteres")
	ErrEmailInvalido    = errors.New("el correo electrónico no es válido")
	ErrPasswordDebil    = errors.New("la contraseña debe tener al menos 8 caracteres")
	ErrPasswordsNoCoinc = errors.New("las contraseñas no coinciden")
	ErrEmailYaExiste    = errors.New("ese correo ya está registrado")
	ErrCredenciales     = errors.New("correo o contraseña incorrectos")
	ErrUltimoAdmin      = errors.New("no se puede quitar el rol al último administrador")
	ErrBorrarseASiMismo = errors.New("no puedes borrarte a ti mismo")
	ErrPasswordActual   = errors.New("la contraseña actual no es correcta")
	ErrPasswordIgual    = errors.New("la contraseña nueva debe ser distinta a la actual")
)

var reEmail = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// UsuarioService ofrece operaciones de negocio sobre usuarios:
// registro, validación y hash de contraseña.
type UsuarioService struct {
	Repo *repository.UsuarioRepo
}

// NuevoUsuarioService construye el servicio.
func NuevoUsuarioService(repo *repository.UsuarioRepo) *UsuarioService {
	return &UsuarioService{Repo: repo}
}

// DatosRegistro agrupa los campos que llegan del formulario.
type DatosRegistro struct {
	Nombre    string
	Email     string
	Password  string
	Password2 string
}

// Registrar valida los datos del formulario, genera el hash de la contraseña
// y guarda el usuario en la BD. Devuelve un error de negocio si algo falla.
func (s *UsuarioService) Registrar(d DatosRegistro) (*models.Usuario, error) {
	nombre := strings.TrimSpace(d.Nombre)
	email := strings.ToLower(strings.TrimSpace(d.Email))

	if n := utf8.RuneCountInString(nombre); n < 2 || n > 60 {
		return nil, ErrNombreInvalido
	}
	if !reEmail.MatchString(email) || len(email) > 120 {
		return nil, ErrEmailInvalido
	}
	if utf8.RuneCountInString(d.Password) < 8 {
		return nil, ErrPasswordDebil
	}
	if d.Password != d.Password2 {
		return nil, ErrPasswordsNoCoinc
	}

	hash, err := HashPassword(d.Password)
	if err != nil {
		return nil, fmt.Errorf("generando hash: %w", err)
	}

	u := &models.Usuario{
		Nombre:       nombre,
		Email:        email,
		PasswordHash: hash,
	}
	if err := s.Repo.Crear(u); err != nil {
		if errors.Is(err, repository.ErrEmailYaRegistrado) {
			return nil, ErrEmailYaExiste
		}
		return nil, err
	}
	return u, nil
}

// HashPassword genera un hash SHA-256 con salt aleatorio de 16 bytes.
// Formato de salida: "salt_hex$hash_hex". Para un sistema en producción se
// usaría bcrypt o argon2id; para esta PEC académica es suficiente.
//
// Exportado (mayúscula) para que el arranque pueda sembrar el admin por
// defecto sin tocar repositorios directamente.
func HashPassword(pwd string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(pwd))
	sum := h.Sum(nil)
	return hex.EncodeToString(salt) + "$" + hex.EncodeToString(sum), nil
}

// VerificarPassword comprueba si la contraseña en claro coincide con el hash guardado.
func VerificarPassword(pwd, almacenado string) bool {
	partes := strings.Split(almacenado, "$")
	if len(partes) != 2 {
		return false
	}
	salt, err := hex.DecodeString(partes[0])
	if err != nil {
		return false
	}
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(pwd))
	return hex.EncodeToString(h.Sum(nil)) == partes[1]
}

// Existe comprueba si un email ya está registrado (utilidad).
func (s *UsuarioService) Existe(email string) (bool, error) {
	_, err := s.Repo.BuscarPorEmail(email)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

// IniciarSesion valida email y contraseña; devuelve el usuario autenticado
// o ErrCredenciales si algo no cuadra.
func (s *UsuarioService) IniciarSesion(email, password string) (*models.Usuario, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return nil, ErrCredenciales
	}
	u, err := s.Repo.BuscarPorEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCredenciales
		}
		return nil, err
	}
	if !VerificarPassword(password, u.PasswordHash) {
		return nil, ErrCredenciales
	}
	return u, nil
}

// Listar devuelve todos los usuarios. Expuesto para el panel admin.
func (s *UsuarioService) Listar() ([]models.Usuario, error) {
	return s.Repo.Listar()
}

// Contar devuelve cuántos usuarios hay registrados.
func (s *UsuarioService) Contar() (int, error) {
	return s.Repo.Contar()
}

// ObtenerPorID expone la búsqueda por id al handler admin.
func (s *UsuarioService) ObtenerPorID(id int64) (*models.Usuario, error) {
	return s.Repo.ObtenerPorID(id)
}

// Borrar elimina un usuario. Salvaguardas:
//   - no se puede borrar uno mismo (actorID == id)
//   - no se puede borrar el último administrador que quede.
func (s *UsuarioService) Borrar(actorID, id int64) error {
	if actorID == id {
		return ErrBorrarseASiMismo
	}
	obj, err := s.Repo.ObtenerPorID(id)
	if err != nil {
		return err
	}
	if obj.EsAdmin {
		nAdmins, err := s.Repo.ContarAdmins()
		if err != nil {
			return err
		}
		if nAdmins <= 1 {
			return ErrUltimoAdmin
		}
	}
	return s.Repo.Borrar(id)
}

// ResetPassword genera una contraseña aleatoria nueva de 12 caracteres,
// actualiza el hash del usuario y la devuelve en claro para mostrarla UNA
// sola vez al admin (el usuario afectado deberá cambiarla al volver a entrar).
func (s *UsuarioService) ResetPassword(id int64) (string, error) {
	nueva, err := generarPasswordAleatoria(12)
	if err != nil {
		return "", err
	}
	hash, err := HashPassword(nueva)
	if err != nil {
		return "", err
	}
	if err := s.Repo.ActualizarPasswordHash(id, hash); err != nil {
		return "", err
	}
	return nueva, nil
}

// CambiarPassword permite a un usuario cambiar su propia contraseña.
// Verifica primero la contraseña actual, exige que la nueva cumpla los
// mismos requisitos que en el registro y que coincida con la confirmación.
func (s *UsuarioService) CambiarPassword(id int64, actual, nueva, nueva2 string) error {
	u, err := s.Repo.ObtenerPorID(id)
	if err != nil {
		return err
	}
	if !VerificarPassword(actual, u.PasswordHash) {
		return ErrPasswordActual
	}
	if utf8.RuneCountInString(nueva) < 8 {
		return ErrPasswordDebil
	}
	if nueva != nueva2 {
		return ErrPasswordsNoCoinc
	}
	if nueva == actual {
		return ErrPasswordIgual
	}
	hash, err := HashPassword(nueva)
	if err != nil {
		return fmt.Errorf("generando hash: %w", err)
	}
	return s.Repo.ActualizarPasswordHash(id, hash)
}

// ToggleAdmin alterna el rol administrador. Si quitar admin dejaría el
// sistema sin administradores, devuelve ErrUltimoAdmin.
func (s *UsuarioService) ToggleAdmin(id int64) (bool, error) {
	u, err := s.Repo.ObtenerPorID(id)
	if err != nil {
		return false, err
	}
	if u.EsAdmin {
		nAdmins, err := s.Repo.ContarAdmins()
		if err != nil {
			return false, err
		}
		if nAdmins <= 1 {
			return true, ErrUltimoAdmin
		}
	}
	nuevo := !u.EsAdmin
	if err := s.Repo.ActualizarEsAdmin(id, nuevo); err != nil {
		return u.EsAdmin, err
	}
	return nuevo, nil
}

// generarPasswordAleatoria produce una cadena alfanumérica segura.
func generarPasswordAleatoria(longitud int) (string, error) {
	const alfabeto = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	b := make([]byte, longitud)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alfabeto[int(b[i])%len(alfabeto)]
	}
	return string(b), nil
}
