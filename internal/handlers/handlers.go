package handlers

import (
	"net/http"

	"skihub/internal/models"
	"skihub/internal/services"
)

// CookieSesion es el nombre de la cookie HttpOnly que transporta el token.
const CookieSesion = "skihub_session"

// App agrupa las dependencias compartidas que usan todos los handlers:
// servicios de negocio y cache de plantillas.
type App struct {
	Plantillas  Cache
	UsuarioSvc  *services.UsuarioService
	EstacionSvc *services.EstacionService
	NoticiaSvc  *services.NoticiaService
	SesionSvc   *services.SesionService
	FavoritoSvc *services.FavoritoService
	PedidoSvc   *services.PedidoService
	// NieveSvc añade datos en directo de pistas vía infonieve.es.
	// Es opcional: si es nil, los handlers /api/nieve/* devuelven 503.
	NieveSvc *services.NieveService
}

// UsuarioActual intenta recuperar el usuario autenticado a partir de la
// cookie de sesión. Devuelve nil si no hay sesión o ha expirado.
func (a *App) UsuarioActual(r *http.Request) *models.Usuario {
	if a.SesionSvc == nil {
		return nil
	}
	c, err := r.Cookie(CookieSesion)
	if err != nil {
		return nil
	}
	u, err := a.SesionSvc.UsuarioDeToken(r.Context(), c.Value)
	if err != nil || u == nil {
		return nil
	}
	return u
}

// EsAdmin devuelve true si el usuario tiene el rol administrador.
func EsAdmin(u *models.Usuario) bool {
	return u != nil && u.EsAdmin
}

// requerirAdmin comprueba la sesión y el rol. Devuelve el usuario si es
// admin; en caso contrario escribe la respuesta adecuada y retorna nil.
// El llamador debe hacer return inmediatamente si recibe nil.
func (a *App) requerirAdmin(w http.ResponseWriter, r *http.Request) *models.Usuario {
	u := a.UsuarioActual(r)
	if u == nil {
		http.Redirect(w, r, "/login?mensaje=Inicia+sesi%C3%B3n+como+administrador", http.StatusSeeOther)
		return nil
	}
	if !EsAdmin(u) {
		a.NotFound(w, r)
		return nil
	}
	return u
}
