package models

import "time"

// Pedido es la cabecera de una compra confirmada (tabla `orders`).
type Pedido struct {
	ID        int64
	UsuarioID int64
	TotalEur  float64
	Estado    string // paid | pending | cancelled
	CreadoEn  time.Time
}

// PedidoItem es una línea del pedido (tabla `order_items`). Los precios
// y nombre de estación se persisten en cada línea como snapshot
// histórico, para que el ticket sea legible aunque el catálogo cambie.
type PedidoItem struct {
	ID             int64
	PedidoID       int64
	EstacionID     int64
	NombreEstacion string
	TipoPase       string // adult | child | senior
	Cantidad       int
	Dias           int
	FechaInicio    time.Time
	FechaFin       time.Time
	PrecioUnitario float64
	LineaTotal     float64
}
