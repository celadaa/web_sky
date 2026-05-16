package handlers

import "net/http"

// Pago responde a GET /pago con la pasarela de pago demo.
// No procesa pagos reales: solo muestra la interfaz de la pasarela.
func (a *App) Pago(w http.ResponseWriter, r *http.Request) {
	render(w, r, a.Plantillas, "pago", map[string]any{
		"Titulo":      "Pasarela de Pago Segura - SnowBreak",
		"Descripcion": "Completa tu reserva de forfaits de forma segura en SnowBreak.",
		"Activa":      "cesta",
		"Usuario":     a.UsuarioActual(r),
	})
}
