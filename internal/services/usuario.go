// Package services implementa la lógica de negocio de Snowbreak.
// Los handlers llaman a los servicios; los servicios llaman a los
// repositorios. Toda operación que toque la BD recibe context.Context.
package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"skihub/internal/models"
	"skihub/internal/repository"
)

// Errores de negocio que el handler traduce a mensajes para el usuario.
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

// bcryptCost es el factor de coste usado por bcrypt. 12 es un valor
// razonable en hardware actual (~250ms por hash). Se puede subir a 14
// para mayor seguridad si se demuestra que la latencia es aceptable.
const bcryptCost = 12

// UsuarioService ofrece operaciones de negocio sobre usuarios.
type UsuarioService struct {
	Repo *repository.UsuarioRepo
}

// NuevoUsuarioService construye el servicio.
func NuevoUsuarioService(repo *repository.UsuarioRepo) *UsuarioService {
	return &UsuarioService{Repo: repo}
}

// DatosRegistro agrupa los campos del formulario de registro.
type DatosRegistro struct {
	Nombre    string
	Email     string
	Password  string
	Password2 string
}

// Registrar valida los datos del formulario, hashea la contraseña con
// bcrypt y persiste el usuario.
func (s *UsuarioService) Registrar(ctx context.Context, d DatosRegistro) (*models.Usuario, error) {
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
	if err := s.Repo.Crear(ctx, u); err != nil {
		if errors.Is(err, repository.ErrEmailYaRegistrado) {
			return nil, ErrEmailYaExiste
		}
		return nil, err
	}
	return u, nil
}

// HashPassword genera un hash bcrypt con el coste configurado.
//
// Exportado para que el bootstrap del admin pueda generar el hash sin
// pasar por el flujo de registro.
func HashPassword(pwd string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// VerificarPassword comprueba si la contraseña en claro coincide con el
// hash bcrypt almacenado. Devuelve false ante cualquier discrepancia,
// incluyendo hashes mal formados.
func VerificarPassword(pwd, almacenado string) bool {
	return bcrypt.CompareHashAndPassword([]byte(almacenado), []byte(pwd)) == nil
}

// Existe comprueba si un email ya está registrado.
func (s *UsuarioService) Existe(ctx context.Context, email string) (bool, error) {
	_, err := s.Repo.BuscarPorEmail(ctx, email)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

// IniciarSesion valida email y contraseña.
func (s *UsuarioService) IniciarSesion(ctx context.Context, email, password string) (*models.Usuario, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return nil, ErrCredenciales
	}
	u, err := s.Repo.BuscarPorEmail(ctx, email)
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

// Listar devuelve todos los usuarios (panel admin).
func (s *UsuarioService) Listar(ctx context.Context) ([]models.Usuario, error) {
	return s.Repo.Listar(ctx)
}

// Contar devuelve el número total de usuarios.
func (s *UsuarioService) Contar(ctx context.Context) (int, error) {
	return s.Repo.Contar(ctx)
}

// ObtenerPorID devuelve la ficha completa.
func (s *UsuarioService) ObtenerPorID(ctx context.Context, id int64) (*models.Usuario, error) {
	return s.Repo.ObtenerPorID(ctx, id)
}

// Borrar elimina un usuario con salvaguardas (no auto-borrado, no
// borrar al último administrador).
func (s *UsuarioService) Borrar(ctx context.Context, actorID, id int64) error {
	if actorID == id {
		return ErrBorrarseASiMismo
	}
	obj, err := s.Repo.ObtenerPorID(ctx, id)
	if err != nil {
		return err
	}
	if obj.EsAdmin {
		nAdmins, err := s.Repo.ContarAdmins(ctx)
		if err != nil {
			return err
		}
		if nAdmins <= 1 {
			return ErrUltimoAdmin
		}
	}
	return s.Repo.Borrar(ctx, id)
}

// ResetPassword genera una contraseña aleatoria nueva, la hashea con
// bcrypt y la guarda. Devuelve la contraseña en claro UNA sola vez al
// admin para que la transmita al usuario.
func (s *UsuarioService) ResetPassword(ctx context.Context, id int64) (string, error) {
	nueva, err := generarPasswordAleatoria(12)
	if err != nil {
		return "", err
	}
	hash, err := HashPassword(nueva)
	if err != nil {
		return "", err
	}
	if err := s.Repo.ActualizarPasswordHash(ctx, id, hash); err != nil {
		return "", err
	}
	return nueva, nil
}

// CambiarPassword permite a un usuario cambiar su contraseña.
func (s *UsuarioService) CambiarPassword(ctx context.Context, id int64, actual, nueva, nueva2 string) error {
	u, err := s.Repo.ObtenerPorID(ctx, id)
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
	return s.Repo.ActualizarPasswordHash(ctx, id, hash)
}

// ActualizarDatos cambia nombre y/o email validando los formatos. No
// permite tocar contraseña ni rol (esos tienen métodos dedicados).
// Devuelve ErrEmailYaExiste si el nuevo email pertenece a otra cuenta.
func (s *UsuarioService) ActualizarDatos(ctx context.Context, id int64, nombre, email string) error {
	nombre = strings.TrimSpace(nombre)
	email = strings.ToLower(strings.TrimSpace(email))

	if n := utf8.RuneCountInString(nombre); n < 2 || n > 60 {
		return ErrNombreInvalido
	}
	if !reEmail.MatchString(email) || len(email) > 120 {
		return ErrEmailInvalido
	}

	if err := s.Repo.ActualizarDatos(ctx, id, nombre, email); err != nil {
		if errors.Is(err, repository.ErrEmailYaRegistrado) {
			return ErrEmailYaExiste
		}
		return err
	}
	return nil
}

// ToggleAdmin alterna el rol administrador.
func (s *UsuarioService) ToggleAdmin(ctx context.Context, id int64) (bool, error) {
	u, err := s.Repo.ObtenerPorID(ctx, id)
	if err != nil {
		return false, err
	}
	if u.EsAdmin {
		nAdmins, err := s.Repo.ContarAdmins(ctx)
		if err != nil {
			return false, err
		}
		if nAdmins <= 1 {
			return true, ErrUltimoAdmin
		}
	}
	nuevo := !u.EsAdmin
	if err := s.Repo.ActualizarEsAdmin(ctx, id, nuevo); err != nil {
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
