// Package db — utilidades de bootstrap de cuentas administrativas.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// AsegurarAdmin garantiza que existe al menos un usuario con rol admin.
// Si ya hay alguno, no hace nada.
// Si no:
//   - si el email indicado existe como usuario normal, lo promociona a admin;
//   - si no existe, crea un nuevo usuario admin con el hash bcrypt indicado.
//
// Devuelve (creado=true) cuando ha tenido que insertar un usuario nuevo.
func AsegurarAdmin(ctx context.Context, bd *sql.DB, email, nombre, passwordHash string) (creado bool, err error) {
	if email == "" || passwordHash == "" {
		return false, errors.New("AsegurarAdmin: email y passwordHash son obligatorios")
	}

	var n int
	if err = bd.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE is_admin = true`,
	).Scan(&n); err != nil {
		return false, fmt.Errorf("contando admins: %w", err)
	}
	if n > 0 {
		return false, nil
	}

	// ¿El email elegido ya existe como usuario normal? Lo promocionamos.
	var id int64
	err = bd.QueryRowContext(ctx,
		`SELECT id FROM users WHERE email = $1`, email,
	).Scan(&id)
	switch {
	case err == nil:
		_, err = bd.ExecContext(ctx,
			`UPDATE users SET is_admin = true, updated_at = NOW() WHERE id = $1`, id)
		return err == nil, err
	case errors.Is(err, sql.ErrNoRows):
		_, err = bd.ExecContext(ctx,
			`INSERT INTO users (name, email, password_hash, is_admin)
			 VALUES ($1, $2, $3, true)`,
			nombre, email, passwordHash,
		)
		return err == nil, err
	default:
		return false, fmt.Errorf("buscando email admin: %w", err)
	}
}
