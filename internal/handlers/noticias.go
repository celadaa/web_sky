package handlers

import (
	"log"
	"net/http"

	"skihub/internal/models"
)

type datosNoticias struct {
	Titulo          string
	Descripcion     string
	Activa          string
	Noticias        []models.Noticia
	Usuario         *models.Usuario
	CategoriaActiva string
	FallbackCache   bool
	SinNoticias     bool
}

// Noticias responde a GET /noticias.
// Acepta ?categoria=nevada|consejos|evento|general|seguridad|material para filtrar.
func (a *App) Noticias(w http.ResponseWriter, r *http.Request) {
	categoria := r.URL.Query().Get("categoria")

	lista, err := a.NoticiaSvc.ListarConFiltro(r.Context(), categoria)
	if err != nil {
		log.Printf("ERROR listar noticias: %v", err)
		render(w, r, a.Plantillas, "noticias", datosNoticias{
			Titulo:          "Noticias de esqui - SnowBreak",
			Descripcion:     "Ultimas noticias de esqui, reportes de nieve, eventos y consejos.",
			Activa:          "noticias",
			CategoriaActiva: categoria,
			SinNoticias:     true,
			Usuario:         a.UsuarioActual(r),
		})
		return
	}

	render(w, r, a.Plantillas, "noticias", datosNoticias{
		Titulo:          "Noticias de esqui - SnowBreak",
		Descripcion:     "Ultimas noticias de esqui, reportes de nieve, eventos y consejos.",
		Activa:          "noticias",
		CategoriaActiva: categoria,
		Noticias:        lista,
		SinNoticias:     len(lista) == 0,
		Usuario:         a.UsuarioActual(r),
	})
}
