package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"skihub/internal/models"
	"skihub/internal/services"
)

// filaUsuarioAdmin adapta un models.Usuario para la vista: añade el nº de
// favoritos y oculta el hash de la contraseña (no se muestra nunca).
type filaUsuarioAdmin struct {
	ID            int64
	Nombre        string
	Email         string
	FechaRegistro string
	NumFavoritos  int
	EsAdmin       bool
	EsActual      bool // true si esta fila corresponde al admin que está viendo el panel
}

type datosAdminUsuarios struct {
	Titulo      string
	Descripcion string
	Activa      string
	Usuarios    []filaUsuarioAdmin
	Total       int
	Mensaje     string
	Error       string
	NuevaPwd    string // contraseña recién generada por reset (se muestra una única vez)
	PwdUsuario  string // a qué usuario pertenece esa nueva contraseña
	Usuario     *models.Usuario
}

type datosAdminUsuario struct {
	Titulo      string
	Descripcion string
	Activa      string
	Detalle     *models.Usuario
	Favoritas   []models.Estacion
	EsActual    bool
	Usuario     *models.Usuario
}

// AdminUsuarios muestra en /admin/usuarios el listado completo.
// Solo visible para usuarios con rol admin.
func (a *App) AdminUsuarios(w http.ResponseWriter, r *http.Request) {
	actual := a.requerirAdmin(w, r)
	if actual == nil {
		return
	}

	lista, err := a.UsuarioSvc.Listar()
	if err != nil {
		log.Printf("ERROR listar usuarios: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}
	conteo := map[int64]int{}
	if a.FavoritoSvc != nil {
		conteo, err = a.FavoritoSvc.ContarPorUsuario()
		if err != nil {
			log.Printf("ERROR contar favoritos: %v", err)
			conteo = map[int64]int{}
		}
	}

	filas := make([]filaUsuarioAdmin, 0, len(lista))
	for _, u := range lista {
		filas = append(filas, filaUsuarioAdmin{
			ID:            u.ID,
			Nombre:        u.Nombre,
			Email:         u.Email,
			FechaRegistro: u.FechaRegistro.Format("02/01/2006 15:04"),
			NumFavoritos:  conteo[u.ID],
			EsAdmin:       u.EsAdmin,
			EsActual:      u.ID == actual.ID,
		})
	}

	render(w, r, a.Plantillas, "admin_usuarios", datosAdminUsuarios{
		Titulo:      "Administración - Usuarios registrados",
		Descripcion: "Panel de administración de SkiHub.",
		Activa:      "admin",
		Usuarios:    filas,
		Total:       len(filas),
		Mensaje:     r.URL.Query().Get("msg"),
		Error:       r.URL.Query().Get("err"),
		NuevaPwd:    r.URL.Query().Get("nueva"),
		PwdUsuario:  r.URL.Query().Get("para"),
		Usuario:     actual,
	})
}

// AdminUsuarioDetalle muestra la ficha individual de un usuario,
// con su lista de estaciones favoritas.
// Ruta: GET /admin/usuario/{id}
func (a *App) AdminUsuarioDetalle(w http.ResponseWriter, r *http.Request) {
	actual := a.requerirAdmin(w, r)
	if actual == nil {
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/admin/usuario/")
	if idStr == "" || strings.Contains(idStr, "/") {
		a.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		a.NotFound(w, r)
		return
	}
	u, err := a.UsuarioSvc.ObtenerPorID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.NotFound(w, r)
			return
		}
		log.Printf("ERROR obtener usuario %d: %v", id, err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}
	favs, err := a.FavoritoSvc.ListarDeUsuario(id)
	if err != nil {
		log.Printf("ERROR listar favoritas de %d: %v", id, err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	render(w, r, a.Plantillas, "admin_usuario", datosAdminUsuario{
		Titulo:      "Administración - " + u.Nombre,
		Descripcion: "Detalle del usuario " + u.Email,
		Activa:      "admin",
		Detalle:     u,
		Favoritas:   favs,
		EsActual:    u.ID == actual.ID,
		Usuario:     actual,
	})
}

// AdminBorrarUsuario borra una cuenta (POST /admin/usuarios/borrar).
func (a *App) AdminBorrarUsuario(w http.ResponseWriter, r *http.Request) {
	actual := a.requerirAdmin(w, r)
	if actual == nil {
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	id, ok := parseIDForm(r, "usuario_id")
	if !ok {
		http.Redirect(w, r, "/admin/usuarios?err=ID+inv%C3%A1lido", http.StatusSeeOther)
		return
	}
	if err := a.UsuarioSvc.Borrar(actual.ID, id); err != nil {
		mensaje := "Error al borrar"
		switch {
		case errors.Is(err, services.ErrBorrarseASiMismo):
			mensaje = "No puedes borrarte a ti mismo"
		case errors.Is(err, services.ErrUltimoAdmin):
			mensaje = "No puedes borrar el último administrador"
		default:
			log.Printf("ERROR borrar usuario %d: %v", id, err)
		}
		http.Redirect(w, r, "/admin/usuarios?err="+urlEncode(mensaje), http.StatusSeeOther)
		return
	}
	log.Printf("Admin %d borró al usuario %d", actual.ID, id)
	http.Redirect(w, r, "/admin/usuarios?msg=Usuario+eliminado", http.StatusSeeOther)
}

// AdminResetPassword genera una contraseña aleatoria para el usuario
// indicado (POST /admin/usuarios/reset).
func (a *App) AdminResetPassword(w http.ResponseWriter, r *http.Request) {
	actual := a.requerirAdmin(w, r)
	if actual == nil {
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	id, ok := parseIDForm(r, "usuario_id")
	if !ok {
		http.Redirect(w, r, "/admin/usuarios?err=ID+inv%C3%A1lido", http.StatusSeeOther)
		return
	}
	u, err := a.UsuarioSvc.ObtenerPorID(id)
	if err != nil {
		http.Redirect(w, r, "/admin/usuarios?err=Usuario+no+encontrado", http.StatusSeeOther)
		return
	}
	nueva, err := a.UsuarioSvc.ResetPassword(id)
	if err != nil {
		log.Printf("ERROR reset password %d: %v", id, err)
		http.Redirect(w, r, "/admin/usuarios?err=Error+al+resetear", http.StatusSeeOther)
		return
	}
	log.Printf("Admin %d reseteó contraseña de usuario %d", actual.ID, id)
	http.Redirect(w, r,
		"/admin/usuarios?nueva="+urlEncode(nueva)+"&para="+urlEncode(u.Email),
		http.StatusSeeOther)
}

// AdminToggleAdmin promueve / revoca el rol admin de un usuario
// (POST /admin/usuarios/toggle-admin).
func (a *App) AdminToggleAdmin(w http.ResponseWriter, r *http.Request) {
	actual := a.requerirAdmin(w, r)
	if actual == nil {
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	id, ok := parseIDForm(r, "usuario_id")
	if !ok {
		http.Redirect(w, r, "/admin/usuarios?err=ID+inv%C3%A1lido", http.StatusSeeOther)
		return
	}
	nuevo, err := a.UsuarioSvc.ToggleAdmin(id)
	if err != nil {
		mensaje := "Error al cambiar rol"
		if errors.Is(err, services.ErrUltimoAdmin) {
			mensaje = "No puedes quitar el rol al último administrador"
		} else {
			log.Printf("ERROR toggle admin %d: %v", id, err)
		}
		http.Redirect(w, r, "/admin/usuarios?err="+urlEncode(mensaje), http.StatusSeeOther)
		return
	}
	log.Printf("Admin %d cambió rol de usuario %d a admin=%v", actual.ID, id, nuevo)
	msg := "Rol actualizado: ahora es usuario"
	if nuevo {
		msg = "Rol actualizado: ahora es administrador"
	}
	http.Redirect(w, r, "/admin/usuarios?msg="+urlEncode(msg), http.StatusSeeOther)
}

// parseIDForm extrae y valida un id entero positivo del formulario.
func parseIDForm(r *http.Request, campo string) (int64, bool) {
	if err := r.ParseForm(); err != nil {
		return 0, false
	}
	id, err := strconv.ParseInt(r.FormValue(campo), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// urlEncode escapa el valor para usarlo en una query string.
func urlEncode(s string) string {
	return url.QueryEscape(s)
}
