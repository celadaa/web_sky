package handlers

import (
	"log"
	"net/http"

	"skihub/internal/models"
)

type datosNoticias struct {
	Titulo      string
	Descripcion string
	Activa      string
	Noticias    []models.Noticia
	Usuario     *models.Usuario
}

// Noticias responde a GET /noticias.
func (a *App) Noticias(w http.ResponseWriter, r *http.Request) {
	lista, err := a.NoticiaSvc.Listar()
	if err != nil {
		log.Printf("ERROR listar noticias: %v", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}
	render(w, r, a.Plantillas, "noticias", datosNoticias{
		Titulo:      "Noticias de esquí - SkiHub",
		Descripcion: "Últimas noticias de esquí, reportes de nieve, eventos y consejos para esquiadores en SkiHub.",
		Activa:      "noticias",
		Noticias:    lista,
		Usuario:     a.UsuarioActual(r),
	})
}
