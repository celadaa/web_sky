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
