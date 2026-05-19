package handlers

import (
	"net/http"

	"skihub/internal/models"
)

// datosPremium agrupa los datos que la landing premium necesita.
// De momento solo el contexto base (usuario, sección activa); los
// productos del market se renderizan en el cliente desde JS para
// poder iterar el diseño sin tocar el backend.
type datosPremium struct {
	Titulo      string
	Descripcion string
	Activa      string
	Usuario     *models.Usuario
}

// Premium responde a GET /premium con la nueva landing brutalista de
// portfolio. La home original ("/") se mantiene intacta. Cuando se quiera
// promocionar esta landing a home, basta con sustituir el handler en
// cmd/servidor/main.go.
func (a *App) Premium(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/premium" {
		a.NotFound(w, r)
		return
	}
	u := a.UsuarioActual(r)
	render(w, r, a.Plantillas, "premium", datosPremium{
		Titulo:      "SnowBreak Market | Alquila · Esquía · Devuelve",
		Descripcion: "Material de esquí premium, presentado como un supermercado de alta gama. Reserva pack, casco, seguro y mantenimiento desde una experiencia editorial.",
		Activa:      "premium",
		Usuario:     u,
	})
}
