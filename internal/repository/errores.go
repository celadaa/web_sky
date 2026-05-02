// Package repository — utilidades comunes a todos los repos.
//
// Aquí centralizamos:
//   - errores de dominio que pueden devolver varios repositorios (Email
//     duplicado, etc.)
//   - helpers para detectar códigos SQLSTATE específicos de Postgres.
package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrEmailYaRegistrado se devuelve al intentar crear un usuario con un
// email que ya existe (violación UNIQUE en users.email).
var ErrEmailYaRegistrado = errors.New("el email ya está registrado")

// Códigos SQLSTATE de PostgreSQL que nos interesan.
// Ver: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgCheckViolation      = "23514"
	pgNotNullViolation    = "23502"
)

// esCodigoPg comprueba si `err` envuelve un *pgconn.PgError con el
// código SQLSTATE indicado.
func esCodigoPg(err error, code string) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == code
	}
	return false
}

// esUnique devuelve true si el error es una violación UNIQUE.
func esUnique(err error) bool { return esCodigoPg(err, pgUniqueViolation) }
