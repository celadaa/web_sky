// Package repository — tablas `orders` y `order_items` de PostgreSQL.
//
// Un pedido se inserta como una unidad: cabecera + N líneas se persisten
// dentro de la misma transacción. Si cualquier línea falla por una
// restricción (CHECK, FK, etc.) se hace ROLLBACK y el pedido no se
// guarda parcialmente.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"skihub/internal/models"
)

// PedidoRepo encapsula las operaciones sobre pedidos y sus líneas.
type PedidoRepo struct {
	BD *sql.DB
}

// NuevoPedidoRepo construye el repositorio.
func NuevoPedidoRepo(bd *sql.DB) *PedidoRepo {
	return &PedidoRepo{BD: bd}
}

// Crear persiste cabecera + líneas en una transacción y devuelve el ID
// del pedido creado. El total se recalcula sumando las líneas para
// evitar inconsistencias provocadas por el cliente.
func (r *PedidoRepo) Crear(ctx context.Context, usuarioID int64, items []models.PedidoItem) (int64, error) {
	if len(items) == 0 {
		return 0, fmt.Errorf("pedido vacío: añade al menos una línea")
	}

	var total float64
	for _, it := range items {
		total += it.LineaTotal
	}

	tx, err := r.BD.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("abrir tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var pedidoID int64
	err = tx.QueryRowContext(ctx,
		`INSERT INTO orders (user_id, total_eur, status)
		 VALUES ($1, $2, 'paid')
		 RETURNING id`,
		usuarioID, total,
	).Scan(&pedidoID)
	if err != nil {
		return 0, fmt.Errorf("insert orders: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO order_items
		  (order_id, station_id, station_name, pass_type,
		   quantity, days, start_date, end_date,
		   unit_price_eur, line_total_eur)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`)
	if err != nil {
		return 0, fmt.Errorf("preparar insert order_items: %w", err)
	}
	defer stmt.Close()

	for i, it := range items {
		if _, err := stmt.ExecContext(ctx,
			pedidoID, it.EstacionID, it.NombreEstacion, it.TipoPase,
			it.Cantidad, it.Dias, it.FechaInicio, it.FechaFin,
			it.PrecioUnitario, it.LineaTotal,
		); err != nil {
			return 0, fmt.Errorf("insert order_item %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit pedido: %w", err)
	}
	return pedidoID, nil
}

// ListarDeUsuario devuelve los pedidos del usuario, más recientes primero
// (cabeceras solamente). Para mostrar la cesta histórica al usuario.
func (r *PedidoRepo) ListarDeUsuario(ctx context.Context, usuarioID int64) ([]models.Pedido, error) {
	rows, err := r.BD.QueryContext(ctx,
		`SELECT id, user_id, total_eur, status, created_at
		 FROM orders
		 WHERE user_id = $1
		 ORDER BY created_at DESC, id DESC`,
		usuarioID,
	)
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	var lista []models.Pedido
	for rows.Next() {
		var p models.Pedido
		if err := rows.Scan(&p.ID, &p.UsuarioID, &p.TotalEur, &p.Estado, &p.CreadoEn); err != nil {
			return nil, err
		}
		lista = append(lista, p)
	}
	return lista, rows.Err()
}
