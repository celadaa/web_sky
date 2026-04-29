package handlers

import (
	"log"
	"net/http"

	"skihub/internal/models"
)

// datosForfaits agrupa la información que la plantilla de forfaits necesita.
// Reutilizamos la lista de estaciones porque las tarjetas se construyen sobre
// los mismos datos (precios, imagen, ubicación...).
type datosForfaits struct {
	Titulo      string
	Descripcion string
	Activa      string
	Estaciones  []models.Estacion
	Usuario     *models.Usuario
}

// Forfaits responde a GET /forfaits con el listado de estaciones disponibles
// para comprar pases de día. La compra se realiza en la ficha de cada estación.
func (a *App) Forfaits(w http.ResponseWriter, r *http.Request) {
	u := a.UsuarioActual(r)
	var uid int64
	if u != nil {
		uid = u.ID
	}
	lista, err := a.EstacionSvc.Listar(uid)
	if err != nil {
		log.Printf("ERROR listar forfaits: %v", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	render(w, r, a.Plantillas, "forfaits", datosForfaits{
		Titulo:      "Comprar Forfaits - SkiHub",
		Descripcion: "Compra forfaits de día para las estaciones de esquí del Pirineo y Sierra Nevada.",
		Activa:      "forfaits",
		Estaciones:  lista,
		Usuario:     u,
	})
}

// datosCesta es la estructura mínima que necesita la plantilla de la cesta.
// La cesta vive en localStorage del navegador, así que en servidor solo
// necesitamos los datos comunes del layout (usuario, navegación activa).
type datosCesta struct {
	Titulo      string
	Descripcion string
	Activa      string
	Usuario     *models.Usuario
}

// Cesta responde a GET /cesta con la página de revisión de la cesta.
// La lógica de presentación de los ítems se hace 100% en JavaScript a partir
// de los datos persistidos en localStorage.
func (a *App) Cesta(w http.ResponseWriter, r *http.Request) {
	render(w, r, a.Plantillas, "cesta", datosCesta{
		Titulo:      "Tu cesta - SkiHub",
		Descripcion: "Revisa los forfaits añadidos a tu cesta de la compra.",
		Activa:      "cesta",
		Usuario:     a.UsuarioActual(r),
	})
}
