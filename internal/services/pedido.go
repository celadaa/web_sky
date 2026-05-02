// Package services — pedidos (forfaits comprados a través del checkout).
//
// La cesta del cliente vive en localStorage del navegador. Al confirmar la
// compra (POST /api/cesta/checkout) el handler manda al servidor un JSON
// con las líneas. El servicio valida cada línea ANTES de tocar la BD:
//   - estación existe
//   - tipo de pase válido
//   - cantidad y días positivos
//   - fechas coherentes
//   - el precio unitario coincide con el catálogo (defensa contra
//     manipulación del cliente: nunca se confía en el precio que envía)
//
// El total se recalcula en servidor a partir de los precios oficiales,
// se ignora el total que pueda mandar el cliente.
package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"skihub/internal/models"
	"skihub/internal/repository"
)

// Errores de negocio del checkout.
var (
	ErrCestaVacia          = errors.New("la cesta está vacía")
	ErrItemEstacion        = errors.New("estación no encontrada")
	ErrItemTipoPase        = errors.New("tipo de pase inválido (adult|child|senior)")
	ErrItemCantidad        = errors.New("la cantidad debe ser mayor que cero")
	ErrItemDias            = errors.New("la duración debe ser de al menos un día")
	ErrItemFechas          = errors.New("la fecha de fin no puede ser anterior a la de inicio")
	ErrItemFechaFormato    = errors.New("formato de fecha inválido (YYYY-MM-DD)")
	ErrItemPrecioInvalido  = errors.New("el precio no coincide con el del catálogo")
)

// LineaCesta es lo que llega del cliente. Validamos cada campo aquí.
type LineaCesta struct {
	EstacionID     int64   `json:"id_estacion"`
	NombreEstacion string  `json:"nombre_estacion"`
	TipoPase       string  `json:"tipo_pase"`       // adulto | nino | senior
	FechaInicio    string  `json:"fecha_inicio"`    // YYYY-MM-DD
	FechaFin       string  `json:"fecha_fin"`       // YYYY-MM-DD
	Cantidad       int     `json:"cantidad"`
	Dias           int     `json:"dias"`            // se ignora y se recalcula
	PrecioUnitario float64 `json:"precio_unitario"` // se valida contra catálogo
}

// PedidoService implementa el checkout.
type PedidoService struct {
	Repo    *repository.PedidoRepo
	EstRepo *repository.EstacionRepo
}

// NuevoPedidoService construye el servicio.
func NuevoPedidoService(repo *repository.PedidoRepo, est *repository.EstacionRepo) *PedidoService {
	return &PedidoService{Repo: repo, EstRepo: est}
}

// ResumenPedido es lo que devolvemos al frontend tras un checkout
// satisfactorio. Es deliberadamente minimalista para no exponer
// internals.
type ResumenPedido struct {
	PedidoID int64   `json:"pedido_id"`
	Total    float64 `json:"total_eur"`
	Lineas   int     `json:"num_lineas"`
}

// Crear valida la cesta y persiste un pedido + sus líneas.
func (s *PedidoService) Crear(ctx context.Context, usuarioID int64, lineas []LineaCesta) (*ResumenPedido, error) {
	if len(lineas) == 0 {
		return nil, ErrCestaVacia
	}

	items := make([]models.PedidoItem, 0, len(lineas))
	var totalServidor float64
	for i, raw := range lineas {
		item, err := s.validarLinea(ctx, raw)
		if err != nil {
			return nil, fmt.Errorf("línea %d: %w", i+1, err)
		}
		items = append(items, item)
		totalServidor += item.LineaTotal
	}

	pedidoID, err := s.Repo.Crear(ctx, usuarioID, items)
	if err != nil {
		return nil, err
	}
	return &ResumenPedido{
		PedidoID: pedidoID,
		Total:    totalServidor,
		Lineas:   len(items),
	}, nil
}

// validarLinea normaliza, valida y enriquece una línea de cesta con los
// datos oficiales del catálogo (nombre y precio reales).
func (s *PedidoService) validarLinea(ctx context.Context, raw LineaCesta) (models.PedidoItem, error) {
	var zero models.PedidoItem

	if raw.Cantidad <= 0 {
		return zero, ErrItemCantidad
	}

	tipoCanon, ok := canonizarTipoPase(raw.TipoPase)
	if !ok {
		return zero, ErrItemTipoPase
	}

	inicio, err := time.Parse("2006-01-02", strings.TrimSpace(raw.FechaInicio))
	if err != nil {
		return zero, ErrItemFechaFormato
	}
	fin, err := time.Parse("2006-01-02", strings.TrimSpace(raw.FechaFin))
	if err != nil {
		return zero, ErrItemFechaFormato
	}
	if fin.Before(inicio) {
		return zero, ErrItemFechas
	}

	dias := int(fin.Sub(inicio).Hours()/24) + 1
	if dias <= 0 {
		return zero, ErrItemDias
	}

	estacion, err := s.EstRepo.ObtenerPorID(ctx, raw.EstacionID)
	if err != nil {
		return zero, ErrItemEstacion
	}

	precioOficial := precioOficialDe(estacion, tipoCanon)

	// Defensa básica contra manipulación: si el cliente miente con
	// el precio rechazamos. Toleramos ±0,01 € por errores de coma flotante.
	if raw.PrecioUnitario > 0 && abs(raw.PrecioUnitario-precioOficial) > 0.01 {
		return zero, ErrItemPrecioInvalido
	}

	lineaTotal := precioOficial * float64(raw.Cantidad) * float64(dias)

	return models.PedidoItem{
		EstacionID:     estacion.ID,
		NombreEstacion: estacion.Nombre,
		TipoPase:       tipoCanon,
		Cantidad:       raw.Cantidad,
		Dias:           dias,
		FechaInicio:    inicio,
		FechaFin:       fin,
		PrecioUnitario: precioOficial,
		LineaTotal:     lineaTotal,
	}, nil
}

// canonizarTipoPase mapea las variantes que envía el frontend a los
// valores que la BD acepta en la constraint CHECK.
func canonizarTipoPase(t string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "adulto", "adult":
		return "adult", true
	case "nino", "niño", "child":
		return "child", true
	case "senior":
		return "senior", true
	default:
		return "", false
	}
}

func precioOficialDe(e *models.Estacion, tipoCanon string) float64 {
	switch tipoCanon {
	case "adult":
		return e.PrecioAdulto
	case "child":
		return e.PrecioNino
	case "senior":
		return e.PrecioSenior
	default:
		return 0
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
