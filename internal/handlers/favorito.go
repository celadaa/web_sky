package handlers

import (
	"log"
	"net/http"
	"strconv"

	"skihub/internal/models"
)

type datosFavoritos struct {
	Titulo      string
	Descripcion string
	Activa      string
	Estaciones  []models.Estacion
	Usuario     *models.Usuario
}

// FavoritosPagina muestra en /favoritos las estaciones marcadas como
// favoritas por el usuario autenticado. Si no hay sesión se redirige a /login.
func (a *App) FavoritosPagina(w http.ResponseWriter, r *http.Request) {
	u := a.UsuarioActual(r)
	if u == nil {
		http.Redirect(w, r, "/login?mensaje=Inicia+sesi%C3%B3n+para+ver+tus+favoritas", http.StatusSeeOther)
		return
	}
	lista, err := a.FavoritoSvc.ListarDeUsuario(u.ID)
	if err != nil {
		log.Printf("ERROR listar favoritos: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}
	render(w, r, a.Plantillas, "favoritos", datosFavoritos{
		Titulo:      "Mis favoritas - SkiHub",
		Descripcion: "Tus estaciones favoritas guardadas en SkiHub.",
		Activa:      "favoritos",
		Estaciones:  lista,
		Usuario:     u,
	})
}

// FavoritoToggle procesa POST /favorito/toggle con el campo "estacion_id".
// Alterna el estado de favorita: si lo era la quita, si no la añade.
// Redirige a la URL indicada en "redirigir" (o a /favoritos por defecto).
func (a *App) FavoritoToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	u := a.UsuarioActual(r)
	if u == nil {
		http.Redirect(w, r, "/login?mensaje=Inicia+sesi%C3%B3n+para+guardar+favoritas", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "datos de formulario incorrectos", http.StatusBadRequest)
		return
	}
	idStr := r.FormValue("estacion_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "estacion_id inválido", http.StatusBadRequest)
		return
	}

	esFav, err := a.FavoritoSvc.Toggle(u.ID, id)
	if err != nil {
		log.Printf("ERROR toggle favorito: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}
	log.Printf("Favorito toggle usuario=%d estacion=%d -> %v", u.ID, id, esFav)

	destino := r.FormValue("redirigir")
	if destino == "" {
		destino = "/favoritos"
	}
	http.Redirect(w, r, destino, http.StatusSeeOther)
}
