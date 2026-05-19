// Package services — lógica de alojamientos cercanos a estaciones y
// reservas de noche.
//
// Decisiones de diseño:
//   - El precio total se calcula SIEMPRE en servidor (precio_noche × noches).
//     Nunca confiamos en el total enviado por el cliente; el cliente solo
//     manda fechas, huéspedes y el ID del alojamiento.
//   - Validaciones cubren: fechas parseables, entrada < salida, fechas no
//     en el pasado, huéspedes en [1,10], máximo 30 noches por reserva.
//   - Si la tabla `lodgings` no existe (migración 012 no aplicada),
//     Listar devuelve lista vacía en vez de error 500 — la sección
//     "Alojamientos cercanos" simplemente no aparece.
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

// Errores tipados para que el handler los mapee a 400 / 404.
var (
	ErrAlojamientoNoEncontrado = errors.New("alojamiento no encontrado")
	ErrFechaFormato            = errors.New("formato de fecha inválido (esperado YYYY-MM-DD)")
	ErrFechaPasado             = errors.New("la fecha de entrada no puede ser en el pasado")
	ErrFechaOrden              = errors.New("la fecha de salida debe ser posterior a la de entrada")
	ErrFechaRangoMax           = errors.New("la reserva no puede superar las 30 noches")
	ErrHuespedes               = errors.New("número de huéspedes debe estar entre 1 y 10")
)

// AlojamientoService orquesta lecturas y reservas.
type AlojamientoService struct {
	Repo *repository.AlojamientoRepo
}

// NuevoAlojamientoService construye el servicio.
func NuevoAlojamientoService(repo *repository.AlojamientoRepo) *AlojamientoService {
	return &AlojamientoService{Repo: repo}
}

// ListarPorEstacion devuelve los alojamientos asociados a la estación.
// Si la tabla aún no existe (migración pendiente) devuelve []  en lugar
// de error: la página de estación seguirá funcionando sin la sección.
func (s *AlojamientoService) ListarPorEstacion(ctx context.Context, estacionID int64) ([]models.Alojamiento, error) {
	if s == nil || s.Repo == nil {
		return nil, nil
	}
	lista, err := s.Repo.ListarPorEstacion(ctx, estacionID)
	if err != nil {
		// "relation lodgings does not exist" → no hemos migrado todavía.
		if strings.Contains(err.Error(), "lodgings") &&
			strings.Contains(strings.ToLower(err.Error()), "does not exist") {
			return nil, nil
		}
		return nil, err
	}
	return lista, nil
}

// Obtener devuelve el detalle de un alojamiento con NombreEstacion.
func (s *AlojamientoService) Obtener(ctx context.Context, id int64) (*models.Alojamiento, error) {
	return s.Repo.ObtenerPorID(ctx, id)
}

// PeticionReserva agrupa los datos que llegan del cliente.
type PeticionReserva struct {
	AlojamientoID int64  `json:"alojamiento_id"`
	FechaEntrada  string `json:"fecha_entrada"` // YYYY-MM-DD
	FechaSalida   string `json:"fecha_salida"`  // YYYY-MM-DD
	Huespedes     int    `json:"huespedes"`
}

// ResumenReserva es lo que devolvemos al cliente tras un POST OK.
type ResumenReserva struct {
	ReservaID    int64   `json:"reserva_id"`
	Noches       int     `json:"noches"`
	TotalEur     float64 `json:"total_eur"`
	NombreAloja  string  `json:"nombre_alojamiento"`
	FechaEntrada string  `json:"fecha_entrada"`
	FechaSalida  string  `json:"fecha_salida"`
}

// Reservar valida la petición, calcula noches y total en servidor, y
// persiste la reserva. Devuelve un resumen para mostrar al usuario.
func (s *AlojamientoService) Reservar(ctx context.Context, usuarioID int64, p PeticionReserva) (*ResumenReserva, error) {
	// 1. Cargar el alojamiento (también valida que existe).
	a, err := s.Repo.ObtenerPorID(ctx, p.AlojamientoID)
	if err != nil {
		return nil, ErrAlojamientoNoEncontrado
	}

	// 2. Validar huéspedes.
	if p.Huespedes < 1 || p.Huespedes > 10 {
		return nil, ErrHuespedes
	}

	// 3. Parsear y validar fechas.
	entrada, err := time.Parse("2006-01-02", p.FechaEntrada)
	if err != nil {
		return nil, ErrFechaFormato
	}
	salida, err := time.Parse("2006-01-02", p.FechaSalida)
	if err != nil {
		return nil, ErrFechaFormato
	}
	hoy := time.Now().Truncate(24 * time.Hour)
	if entrada.Before(hoy) {
		return nil, ErrFechaPasado
	}
	if !salida.After(entrada) {
		return nil, ErrFechaOrden
	}

	// 4. Calcular noches y total en SERVIDOR.
	noches := int(salida.Sub(entrada).Hours() / 24)
	if noches < 1 {
		return nil, ErrFechaOrden
	}
	if noches > 30 {
		return nil, ErrFechaRangoMax
	}
	total := a.PrecioNoche * float64(noches)

	// 5. Persistir.
	r := &models.ReservaAlojamiento{
		UsuarioID:      usuarioID,
		AlojamientoID:  a.ID,
		NombreAloja:    a.Nombre,
		TipoAloja:      a.Tipo,
		EstacionID:     a.EstacionID,
		NombreEstacion: a.NombreEstacion,
		FechaEntrada:   entrada,
		FechaSalida:    salida,
		Noches:         noches,
		Huespedes:      p.Huespedes,
		PrecioNoche:    a.PrecioNoche,
		TotalEur:       total,
		Estado:         "confirmed",
	}
	id, err := s.Repo.InsertarReserva(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("persistir reserva: %w", err)
	}
	return &ResumenReserva{
		ReservaID:    id,
		Noches:       noches,
		TotalEur:     total,
		NombreAloja:  a.Nombre,
		FechaEntrada: p.FechaEntrada,
		FechaSalida:  p.FechaSalida,
	}, nil
}
