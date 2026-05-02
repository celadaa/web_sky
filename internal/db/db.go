// Package db gestiona la conexión a PostgreSQL y la aplicación de migraciones.
//
// Se usa el driver `pgx` (github.com/jackc/pgx/v5/stdlib) registrado como
// "pgx" en el paquete database/sql. Esto nos permite mantener las firmas
// existentes de los repositorios (*sql.DB, sql.Rows, sql.ErrNoRows) y, al
// mismo tiempo, beneficiarnos del rendimiento y los errores tipados de
// pgx (`pgconn.PgError`).
//
// Las migraciones son ficheros .sql en el directorio db/migrations,
// numerados con tres dígitos (001_*, 002_*, ...). El runner aplica los
// que aún no constan en la tabla `schema_migrations` en orden numérico,
// dentro de una transacción por fichero. Es **idempotente**: ejecutarlo
// dos veces no aplica nada la segunda.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"skihub/internal/config"

	// Driver pgx registrado como "pgx" en database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Conectar abre el pool de conexiones a PostgreSQL con los parámetros del
// Config y comprueba que el servidor responde antes de devolver el handle.
//
// Devuelve un *sql.DB pensado para ser inyectado en los repositorios.
func Conectar(cfg *config.Config) (*sql.DB, error) {
	bd, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("abriendo conexión Postgres: %w", err)
	}
	bd.SetMaxOpenConns(cfg.DBMaxOpenConns)
	bd.SetMaxIdleConns(cfg.DBMaxIdleConns)
	bd.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := bd.PingContext(ctx); err != nil {
		_ = bd.Close()
		return nil, fmt.Errorf("ping Postgres %s:%d: %w", cfg.DBHost, cfg.DBPort, err)
	}
	log.Printf("Conectado a PostgreSQL %s@%s:%d/%s (sslmode=%s)",
		cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
	return bd, nil
}

// EjecutarMigraciones aplica los ficheros SQL de `dir` que aún no hayan
// sido registrados en la tabla `schema_migrations`. Cada fichero se
// ejecuta dentro de su propia transacción.
//
// Convención de nombre: NNN_descripcion.sql (NNN = entero con padding).
// Ejemplo: 001_create_users.sql, 010_seed_news.sql.
func EjecutarMigraciones(ctx context.Context, bd *sql.DB, dir string) error {
	if err := asegurarTablaMigraciones(ctx, bd); err != nil {
		return err
	}
	aplicadas, err := leerAplicadas(ctx, bd)
	if err != nil {
		return err
	}

	ficheros, err := listarMigraciones(dir)
	if err != nil {
		return err
	}
	if len(ficheros) == 0 {
		log.Printf("[migraciones] no se encontraron ficheros .sql en %s", dir)
		return nil
	}

	pendientes := 0
	for _, f := range ficheros {
		nombre := filepath.Base(f)
		if aplicadas[nombre] {
			continue
		}
		if err := aplicarFichero(ctx, bd, f, nombre); err != nil {
			return fmt.Errorf("migración %s: %w", nombre, err)
		}
		log.Printf("[migraciones] ✓ aplicada: %s", nombre)
		pendientes++
	}
	if pendientes == 0 {
		log.Printf("[migraciones] esquema al día (%d ya aplicadas)", len(aplicadas))
	} else {
		log.Printf("[migraciones] aplicadas %d nuevas migraciones", pendientes)
	}
	return nil
}

// asegurarTablaMigraciones crea la tabla de control si no existe. La tabla
// es deliberadamente sencilla: filename + cuándo se aplicó.
func asegurarTablaMigraciones(ctx context.Context, bd *sql.DB) error {
	_, err := bd.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename   TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("creando schema_migrations: %w", err)
	}
	return nil
}

// leerAplicadas devuelve un set con los nombres de fichero ya registrados.
func leerAplicadas(ctx context.Context, bd *sql.DB) (map[string]bool, error) {
	rows, err := bd.QueryContext(ctx, `SELECT filename FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("leyendo schema_migrations: %w", err)
	}
	defer rows.Close()
	set := map[string]bool{}
	for rows.Next() {
		var f string
		if err := rows.Scan(&f); err != nil {
			return nil, err
		}
		set[f] = true
	}
	return set, rows.Err()
}

// listarMigraciones devuelve los ficheros .sql del directorio ordenados
// por nombre (lo que efectivamente los ordena por número, gracias al padding).
func listarMigraciones(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no existe el directorio de migraciones %q (créalo y añade ficheros .sql)", dir)
		}
		return nil, fmt.Errorf("listando %s: %w", dir, err)
	}
	var lista []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		lista = append(lista, filepath.Join(dir, e.Name()))
	}
	sort.Strings(lista)
	return lista, nil
}

// aplicarFichero ejecuta el contenido de un .sql en una transacción y, si
// va bien, lo registra en schema_migrations. Si falla, hace ROLLBACK.
func aplicarFichero(ctx context.Context, bd *sql.DB, ruta, nombre string) error {
	contenido, err := os.ReadFile(ruta)
	if err != nil {
		return fmt.Errorf("leyendo %s: %w", ruta, err)
	}
	tx, err := bd.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, string(contenido)); err != nil {
		return fmt.Errorf("ejecutando: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations(filename) VALUES ($1)`, nombre,
	); err != nil {
		return fmt.Errorf("registrando en schema_migrations: %w", err)
	}
	return tx.Commit()
}
