// Package handlers — endpoint de checkout de la cesta de la compra.
//
//	POST /api/cesta/checkout
//
// Body JSON aceptado:
//
//	{
//	  "items": [
//	    {
//	      "id_estacion":     12,
//	      "nombre_estacion": "Sierra Nevada",
//	      "tipo_pase":       "adulto",
//	      "fecha_inicio":    "2026-12-15",
//	      "fecha_fin":       "2026-12-17",
//	      "cantidad":        2,
//	      "precio_unitario": 58.0
//	    }, ...
//	  ]
//	}
//
// Respuestas:
//
//	200 → { "pedido_id": 42, "total_eur": 348.00, "num_lineas": 1 }
//	400 → datos inválidos (formato, cantidades, fechas, precio…)
//	401 → sin sesión
//	500 → error interno
//
// El precio enviado por el cliente se valida contra el catálogo:
// nunca confiamos en él. El total se recalcula en servidor.
package handlers

import (
	"errors"
	"log"
	"net/http"

	"skihub/internal/services"
)

type peticionCheckout struct {
	Items []services.LineaCesta `json:"items"`
}

// ApiCheckout responde a POST /api/cesta/checkout.
func (a *App) ApiCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		escribirError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}
	if a.PedidoSvc == nil {
		escribirError(w, http.StatusServiceUnavailable, "servicio de pedidos no disponible")
		return
	}

	u := a.UsuarioActual(r)
	if u == nil {
		escribirError(w, http.StatusUnauthorized, "inicia sesión para confirmar la compra")
		return
	}

	var p peticionCheckout
	if err := decodificarJSON(r, &p); err != nil {
		escribirError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	resumen, err := a.PedidoSvc.Crear(r.Context(), u.ID, p.Items)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCestaVacia),
			errors.Is(err, services.ErrItemEstacion),
			errors.Is(err, services.ErrItemTipoPase),
			errors.Is(err, services.ErrItemCantidad),
			errors.Is(err, services.ErrItemDias),
			errors.Is(err, services.ErrItemFechas),
			errors.Is(err, services.ErrItemFechaFormato),
			errors.Is(err, services.ErrItemPrecioInvalido):
			escribirError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("ERROR checkout usuario=%d: %v", u.ID, err)
			escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		}
		return
	}
	log.Printf("Checkout OK usuario=%d pedido=%d total=%.2f€",
		u.ID, resumen.PedidoID, resumen.Total)
	escribirJSON(w, http.StatusOK, resumen)
}
