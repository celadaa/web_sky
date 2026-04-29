// Package repository encapsula todo el acceso a la base de datos.
// Los servicios usan estos repositorios, nunca hablan directamente con database/sql.
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"skihub/internal/models"
)

// ErrEmailYaRegistrado se devuelve cuando se intenta crear un usuario
// con un email que ya existe (UNIQUE violation en SQLite).
var ErrEmailYaRegistrado = errors.New("el email ya está registrado")

// UsuarioRepo ofrece operaciones CRUD sobre la tabla `usuarios`.
type UsuarioRepo struct {
	BD *sql.DB
}

// NuevoUsuarioRepo construye un repositorio listo para usar.
func NuevoUsuarioRepo(bd *sql.DB) *UsuarioRepo {
	return &UsuarioRepo{BD: bd}
}

// Crear inserta un nuevo usuario. Devuelve ErrEmailYaRegistrado si el
// email ya existe, o cualquier otro error de la BD.
func (r *UsuarioRepo) Crear(u *models.Usuario) error {
	esAdmin := 0
	if u.EsAdmin {
		esAdmin = 1
	}
	res, err := r.BD.Exec(
		`INSERT INTO usuarios (nombre, email, password_hash, es_admin)
		 VALUES (?, ?, ?, ?)`,
		u.Nombre, u.Email, u.PasswordHash, esAdmin,
	)
	if err != nil {
		if esErrorUnique(err) {
			return ErrEmailYaRegistrado
		}
		return fmt.Errorf("insert usuario: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = id
	return nil
}

// BuscarPorEmail devuelve el usuario con ese email, o sql.ErrNoRows si no existe.
func (r *UsuarioRepo) BuscarPorEmail(email string) (*models.Usuario, error) {
	u := &models.Usuario{}
	var esAdmin int
	err := r.BD.QueryRow(
		`SELECT id, nombre, email, password_hash, fecha_registro, es_admin
		 FROM usuarios WHERE email = ?`, email,
	).Scan(&u.ID, &u.Nombre, &u.Email, &u.PasswordHash, &u.FechaRegistro, &esAdmin)
	if err != nil {
		return nil, err
	}
	u.EsAdmin = esAdmin != 0
	return u, nil
}

// ObtenerPorID devuelve el usuario con ese id, o sql.ErrNoRows si no existe.
func (r *UsuarioRepo) ObtenerPorID(id int64) (*models.Usuario, error) {
	u := &models.Usuario{}
	var esAdmin int
	err := r.BD.QueryRow(
		`SELECT id, nombre, email, password_hash, fecha_registro, es_admin
		 FROM usuarios WHERE id = ?`, id,
	).Scan(&u.ID, &u.Nombre, &u.Email, &u.PasswordHash, &u.FechaRegistro, &esAdmin)
	if err != nil {
		return nil, err
	}
	u.EsAdmin = esAdmin != 0
	return u, nil
}

// Listar devuelve todos los usuarios registrados, ordenados por fecha de
// registro descendente (los más recientes primero).
func (r *UsuarioRepo) Listar() ([]models.Usuario, error) {
	rows, err := r.BD.Query(
		`SELECT id, nombre, email, password_hash, fecha_registro, es_admin
		 FROM usuarios
		 ORDER BY fecha_registro DESC, id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query usuarios: %w", err)
	}
	defer rows.Close()
	var lista []models.Usuario
	for rows.Next() {
		var u models.Usuario
		var esAdmin int
		if err := rows.Scan(&u.ID, &u.Nombre, &u.Email, &u.PasswordHash, &u.FechaRegistro, &esAdmin); err != nil {
			return nil, err
		}
		u.EsAdmin = esAdmin != 0
		lista = append(lista, u)
	}
	return lista, rows.Err()
}

// Contar devuelve el número total de usuarios registrados.
func (r *UsuarioRepo) Contar() (int, error) {
	var n int
	err := r.BD.QueryRow(`SELECT COUNT(*) FROM usuarios`).Scan(&n)
	return n, err
}

// ContarAdmins devuelve cuántos usuarios tienen el rol admin.
func (r *UsuarioRepo) ContarAdmins() (int, error) {
	var n int
	err := r.BD.QueryRow(`SELECT COUNT(*) FROM usuarios WHERE es_admin = 1`).Scan(&n)
	return n, err
}

// Borrar elimina un usuario por su id. Las tablas `sesiones` y `favoritos`
// tienen ON DELETE CASCADE, así que no hay que borrarlas aparte.
func (r *UsuarioRepo) Borrar(id int64) error {
	_, err := r.BD.Exec(`DELETE FROM usuarios WHERE id = ?`, id)
	return err
}

// ActualizarPasswordHash cambia el hash de la contraseña de un usuario.
func (r *UsuarioRepo) ActualizarPasswordHash(id int64, hash string) error {
	_, err := r.BD.Exec(`UPDATE usuarios SET password_hash = ? WHERE id = ?`, hash, id)
	return err
}

// ActualizarEsAdmin fija el rol admin del usuario (true/false).
func (r *UsuarioRepo) ActualizarEsAdmin(id int64, esAdmin bool) error {
	v := 0
	if esAdmin {
		v = 1
	}
	_, err := r.BD.Exec(`UPDATE usuarios SET es_admin = ? WHERE id = ?`, v, id)
	return err
}

// esErrorUnique detecta si el error procede de una violación UNIQUE
// en SQLite (el driver modernc.org/sqlite la reporta como texto).
func esErrorUnique(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint") ||
		strings.Contains(msg, "constraint failed: usuarios.email")
}
