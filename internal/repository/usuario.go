// Package repository — acceso a la tabla `users` de PostgreSQL.
//
// Nota: en SQL la tabla está en inglés (`users`, `name`, `email`,
// `password_hash`, `is_admin`, `created_at`, `updated_at`).
// En Go mantenemos los nombres en español (Nombre, Email, etc.) para
// no romper plantillas ni handlers ya escritos. La traducción se hace
// en este fichero.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"skihub/internal/models"
)

// UsuarioRepo encapsula las operaciones CRUD sobre `users`.
type UsuarioRepo struct {
	BD *sql.DB
}

// NuevoUsuarioRepo construye el repositorio.
func NuevoUsuarioRepo(bd *sql.DB) *UsuarioRepo {
	return &UsuarioRepo{BD: bd}
}

// Crear inserta un nuevo usuario y rellena u.ID con la PK generada
// (vía `RETURNING`). Devuelve ErrEmailYaRegistrado si el email ya
// existe en la tabla.
func (r *UsuarioRepo) Crear(ctx context.Context, u *models.Usuario) error {
	err := r.BD.QueryRowContext(ctx,
		`INSERT INTO users (name, email, password_hash, is_admin)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		u.Nombre, u.Email, u.PasswordHash, u.EsAdmin,
	).Scan(&u.ID, &u.FechaRegistro)
	if err != nil {
		if esUnique(err) {
			return ErrEmailYaRegistrado
		}
		return fmt.Errorf("insert usuario: %w", err)
	}
	return nil
}

// BuscarPorEmail devuelve el usuario con ese email o sql.ErrNoRows.
func (r *UsuarioRepo) BuscarPorEmail(ctx context.Context, email string) (*models.Usuario, error) {
	u := &models.Usuario{}
	err := r.BD.QueryRowContext(ctx,
		`SELECT id, name, email, password_hash, created_at, is_admin
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Nombre, &u.Email, &u.PasswordHash, &u.FechaRegistro, &u.EsAdmin)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// ObtenerPorID devuelve el usuario con ese id o sql.ErrNoRows.
func (r *UsuarioRepo) ObtenerPorID(ctx context.Context, id int64) (*models.Usuario, error) {
	u := &models.Usuario{}
	err := r.BD.QueryRowContext(ctx,
		`SELECT id, name, email, password_hash, created_at, is_admin
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Nombre, &u.Email, &u.PasswordHash, &u.FechaRegistro, &u.EsAdmin)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Listar devuelve todos los usuarios, los más recientes primero.
func (r *UsuarioRepo) Listar(ctx context.Context) ([]models.Usuario, error) {
	rows, err := r.BD.QueryContext(ctx,
		`SELECT id, name, email, password_hash, created_at, is_admin
		 FROM users
		 ORDER BY created_at DESC, id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var lista []models.Usuario
	for rows.Next() {
		var u models.Usuario
		if err := rows.Scan(&u.ID, &u.Nombre, &u.Email, &u.PasswordHash,
			&u.FechaRegistro, &u.EsAdmin); err != nil {
			return nil, err
		}
		lista = append(lista, u)
	}
	return lista, rows.Err()
}

// Contar devuelve el número total de usuarios registrados.
func (r *UsuarioRepo) Contar(ctx context.Context) (int, error) {
	var n int
	err := r.BD.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

// ContarAdmins devuelve el número de usuarios con is_admin = true.
func (r *UsuarioRepo) ContarAdmins(ctx context.Context) (int, error) {
	var n int
	err := r.BD.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE is_admin = true`,
	).Scan(&n)
	return n, err
}

// Borrar elimina un usuario. Las tablas sessions y favorites se limpian
// solas porque sus FKs son ON DELETE CASCADE.
func (r *UsuarioRepo) Borrar(ctx context.Context, id int64) error {
	_, err := r.BD.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

// ActualizarPasswordHash sustituye el hash de la contraseña.
func (r *UsuarioRepo) ActualizarPasswordHash(ctx context.Context, id int64, hash string) error {
	_, err := r.BD.ExecContext(ctx,
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`,
		hash, id)
	return err
}

// ActualizarEsAdmin fija el rol admin del usuario.
func (r *UsuarioRepo) ActualizarEsAdmin(ctx context.Context, id int64, esAdmin bool) error {
	_, err := r.BD.ExecContext(ctx,
		`UPDATE users SET is_admin = $1, updated_at = NOW() WHERE id = $2`,
		esAdmin, id)
	return err
}

// ActualizarDatos cambia nombre y/o email del usuario. Devuelve
// ErrEmailYaRegistrado si el nuevo email choca con otro usuario.
func (r *UsuarioRepo) ActualizarDatos(ctx context.Context, id int64, nombre, email string) error {
	_, err := r.BD.ExecContext(ctx,
		`UPDATE users SET name = $1, email = $2, updated_at = NOW()
		 WHERE id = $3`,
		nombre, email, id)
	if err != nil {
		if esUnique(err) {
			return ErrEmailYaRegistrado
		}
		return fmt.Errorf("update users: %w", err)
	}
	return nil
}

// ============================================================
// Migración 012 — OAuth + verificación de email
// ============================================================

// Las consultas de arriba (Crear, BuscarPorEmail, ObtenerPorID, Listar)
// se mantienen iguales para no romper el código que no necesita los
// campos nuevos. Cuando se requieran (login, callback Google, confirmar
// email), se usan las funciones de abajo que sí escanean las columnas
// añadidas en la migración 012.

// scanCompleto rellena un *models.Usuario a partir de un *sql.Row que
// devuelve TODOS los campos en este orden estable:
//
//	id, name, email, password_hash, created_at, is_admin,
//	email_verified, email_verified_at, google_id, avatar_url,
//	auth_provider, verification_token_hash, verification_token_expires_at
//
// La centralizamos para no repetir el listado en cada query.
func scanCompleto(row interface{ Scan(...any) error }, u *models.Usuario) error {
	return row.Scan(
		&u.ID, &u.Nombre, &u.Email, &u.PasswordHash, &u.FechaRegistro, &u.EsAdmin,
		&u.EmailVerificado, &u.EmailVerificadoEn,
		&u.GoogleID, &u.AvatarURL, &u.AuthProvider,
		&u.VerificationTokenHash, &u.VerificationTokenExpiresAt,
	)
}

const camposCompletos = `
	id, name, email, password_hash, created_at, is_admin,
	email_verified, email_verified_at, google_id, avatar_url,
	auth_provider, verification_token_hash, verification_token_expires_at
`

// BuscarPorEmailCompleto es la versión "con campos OAuth + verificación"
// de BuscarPorEmail. Usada por el flujo de Google login y por verificar email.
func (r *UsuarioRepo) BuscarPorEmailCompleto(ctx context.Context, email string) (*models.Usuario, error) {
	u := &models.Usuario{}
	row := r.BD.QueryRowContext(ctx,
		`SELECT `+camposCompletos+` FROM users WHERE email = $1`, email)
	if err := scanCompleto(row, u); err != nil {
		return nil, err
	}
	return u, nil
}

// ObtenerPorIDCompleto idem, por ID.
func (r *UsuarioRepo) ObtenerPorIDCompleto(ctx context.Context, id int64) (*models.Usuario, error) {
	u := &models.Usuario{}
	row := r.BD.QueryRowContext(ctx,
		`SELECT `+camposCompletos+` FROM users WHERE id = $1`, id)
	if err := scanCompleto(row, u); err != nil {
		return nil, err
	}
	return u, nil
}

// BuscarPorGoogleID busca un usuario por su sub de Google. sql.ErrNoRows
// si no hay ninguno con ese google_id.
func (r *UsuarioRepo) BuscarPorGoogleID(ctx context.Context, googleID string) (*models.Usuario, error) {
	u := &models.Usuario{}
	row := r.BD.QueryRowContext(ctx,
		`SELECT `+camposCompletos+` FROM users WHERE google_id = $1`, googleID)
	if err := scanCompleto(row, u); err != nil {
		return nil, err
	}
	return u, nil
}

// BuscarPorTokenVerificacion devuelve el usuario cuyo verification_token_hash
// coincide con `hash` y cuyo token NO ha expirado todavía. Si el token no
// existe o caducó devuelve sql.ErrNoRows.
func (r *UsuarioRepo) BuscarPorTokenVerificacion(ctx context.Context, hash string) (*models.Usuario, error) {
	u := &models.Usuario{}
	row := r.BD.QueryRowContext(ctx,
		`SELECT `+camposCompletos+`
		 FROM users
		 WHERE verification_token_hash = $1
		   AND verification_token_expires_at > NOW()`, hash)
	if err := scanCompleto(row, u); err != nil {
		return nil, err
	}
	return u, nil
}

// GuardarTokenVerificacion almacena el hash y la fecha de expiración de
// un token recién generado. Sustituye cualquier token previo del mismo
// usuario (si pide reenvío, el anterior queda invalidado).
func (r *UsuarioRepo) GuardarTokenVerificacion(ctx context.Context, id int64, hash string, expiraEn time.Time) error {
	_, err := r.BD.ExecContext(ctx,
		`UPDATE users
		   SET verification_token_hash       = $1,
		       verification_token_expires_at = $2,
		       updated_at = NOW()
		 WHERE id = $3`,
		hash, expiraEn, id)
	return err
}

// MarcarEmailVerificado pone email_verified=TRUE, email_verified_at=NOW(),
// y limpia el token. Idempotente: si ya está verificado no es error volver
// a llamarla, simplemente refresca el timestamp.
func (r *UsuarioRepo) MarcarEmailVerificado(ctx context.Context, id int64) error {
	_, err := r.BD.ExecContext(ctx,
		`UPDATE users
		   SET email_verified              = TRUE,
		       email_verified_at           = NOW(),
		       verification_token_hash     = NULL,
		       verification_token_expires_at = NULL,
		       updated_at                  = NOW()
		 WHERE id = $1`, id)
	return err
}

// VincularGoogle asocia un google_id (y opcionalmente avatar_url) a un
// usuario ya existente. NO toca password_hash ni auth_provider: el
// usuario sigue pudiendo entrar por email/password si así lo prefiere.
// Devuelve ErrEmailYaRegistrado si el google_id ya está vinculado a otra
// cuenta (violación del UNIQUE INDEX parcial).
func (r *UsuarioRepo) VincularGoogle(ctx context.Context, id int64, googleID string, avatarURL *string) error {
	_, err := r.BD.ExecContext(ctx,
		`UPDATE users
		   SET google_id  = $1,
		       avatar_url = COALESCE($2, avatar_url),
		       updated_at = NOW()
		 WHERE id = $3`, googleID, avatarURL, id)
	if err != nil {
		if esUnique(err) {
			return ErrEmailYaRegistrado
		}
		return fmt.Errorf("vincular google: %w", err)
	}
	return nil
}

// CrearGoogle crea un usuario que se registra por primera vez vía Google.
// Marca auth_provider='google', email_verified al valor que devuelva
// Google y deja password_hash como una cadena no usable ("!google",
// nunca pasará por bcrypt.CompareHashAndPassword con éxito).
//
// Devuelve ErrEmailYaRegistrado si el email ya pertenece a otro usuario
// — el handler debe haber comprobado antes y, si procede, llamar a
// VincularGoogle en su lugar.
func (r *UsuarioRepo) CrearGoogle(ctx context.Context, u *models.Usuario) error {
	// Marca de password "no usable". bcrypt nunca produce un hash con este
	// formato, por lo que VerificarPassword devolverá false sin posibilidad
	// de colisión. Quien quiera entrar con email/password debe pasar antes
	// por un flujo "recuperar contraseña" (futuro).
	if u.PasswordHash == "" {
		u.PasswordHash = "!google"
	}
	if u.AuthProvider == "" {
		u.AuthProvider = "google"
	}
	err := r.BD.QueryRowContext(ctx,
		`INSERT INTO users (
			name, email, password_hash, is_admin,
			email_verified, email_verified_at,
			google_id, avatar_url, auth_provider
		 ) VALUES (
			$1, $2, $3, $4,
			$5, CASE WHEN $5 THEN NOW() ELSE NULL END,
			$6, $7, $8
		 )
		 RETURNING id, created_at`,
		u.Nombre, u.Email, u.PasswordHash, u.EsAdmin,
		u.EmailVerificado,
		u.GoogleID, u.AvatarURL, u.AuthProvider,
	).Scan(&u.ID, &u.FechaRegistro)
	if err != nil {
		if esUnique(err) {
			return ErrEmailYaRegistrado
		}
		return fmt.Errorf("insert usuario google: %w", err)
	}
	return nil
}
