// Package handlers — API REST de usuarios.
//
// Endpoints:
//
//	GET    /api/usuarios       → listar todos los usuarios
//	GET    /api/usuarios/{id}  → obtener un usuario por id
//	POST   /api/usuarios       → crear usuario nuevo (201 + Location)
//	PUT    /api/usuarios/{id}  → actualizar nombre/email
//	DELETE /api/usuarios/{id}  → eliminar usuario y cascada
//
// Todas las respuestas son application/json. Los códigos usados son:
// 200, 201, 400, 401, 403, 404, 405, 409, 500.
package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"skihub/internal/models"
	"skihub/internal/services"
)

// usuarioJSON es la representación pública de un usuario en la API.
// NUNCA incluye el hash de la contraseña ni datos sensibles.
type usuarioJSON struct {
	ID            int64  `json:"id"`
	Nombre        string `json:"nombre"`
	Email         string `json:"email"`
	FechaRegistro string `json:"fecha_registro"`
	EsAdmin       bool   `json:"es_admin"`
}

type peticionCrearUsuario struct {
	Nombre    string `json:"nombre"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Password2 string `json:"password2,omitempty"`
}

type peticionEditarUsuario struct {
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

// respuestaError es el formato uniforme de los mensajes de error JSON.
type respuestaError struct {
	Error string `json:"error"`
}

// aUsuarioJSON convierte un *models.Usuario en su versión segura para
// JSON. Centralizado para garantizar que jamás se filtra el hash.
func aUsuarioJSON(u *models.Usuario) usuarioJSON {
	return usuarioJSON{
		ID:            u.ID,
		Nombre:        u.Nombre,
		Email:         u.Email,
		FechaRegistro: u.FechaRegistro.Format("2006-01-02 15:04:05"),
		EsAdmin:       u.EsAdmin,
	}
}

// escribirJSON serializa el valor con el código HTTP indicado.
func escribirJSON(w http.ResponseWriter, codigo int, datos any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	if datos == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(datos); err != nil {
		log.Printf("ERROR codificando JSON: %v", err)
	}
}

// escribirError envía una respuesta JSON {"error":"..."}.
func escribirError(w http.ResponseWriter, codigo int, msg string) {
	escribirJSON(w, codigo, respuestaError{Error: msg})
}

// requerirAdminAPI: 401 sin sesión, 403 sin rol; nil = aborta.
func (a *App) requerirAdminAPI(w http.ResponseWriter, r *http.Request) *models.Usuario {
	u := a.UsuarioActual(r)
	if u == nil {
		escribirError(w, http.StatusUnauthorized, "no autenticado")
		return nil
	}
	if !EsAdmin(u) {
		escribirError(w, http.StatusForbidden, "se requiere rol de administrador")
		return nil
	}
	return u
}

// ApiUsuarios maneja la colección /api/usuarios.
func (a *App) ApiUsuarios(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.apiListarUsuarios(w, r)
	case http.MethodPost:
		if a.requerirAdminAPI(w, r) == nil {
			return
		}
		a.apiCrearUsuario(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		escribirError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}

// ApiUsuario maneja /api/usuarios/{id}.
func (a *App) ApiUsuario(w http.ResponseWriter, r *http.Request) {
	id, ok := extraerIDRuta(r.URL.Path, "/api/usuarios/")
	if !ok {
		escribirError(w, http.StatusBadRequest, "id no válido")
		return
	}

	switch r.Method {
	case http.MethodGet:
		a.apiObtenerUsuario(w, r, id)
	case http.MethodPut:
		if a.requerirAdminAPI(w, r) == nil {
			return
		}
		a.apiActualizarUsuario(w, r, id)
	case http.MethodDelete:
		actual := a.requerirAdminAPI(w, r)
		if actual == nil {
			return
		}
		a.apiEliminarUsuario(w, r, actual, id)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		escribirError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}

func (a *App) apiListarUsuarios(w http.ResponseWriter, r *http.Request) {
	lista, err := a.UsuarioSvc.Listar(r.Context())
	if err != nil {
		log.Printf("ERROR API listar usuarios: %v", err)
		escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}
	out := make([]usuarioJSON, 0, len(lista))
	for i := range lista {
		out = append(out, aUsuarioJSON(&lista[i]))
	}
	escribirJSON(w, http.StatusOK, out)
}

func (a *App) apiObtenerUsuario(w http.ResponseWriter, r *http.Request, id int64) {
	u, err := a.UsuarioSvc.ObtenerPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			escribirError(w, http.StatusNotFound, "usuario no encontrado")
			return
		}
		log.Printf("ERROR API obtener usuario %d: %v", id, err)
		escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}
	escribirJSON(w, http.StatusOK, aUsuarioJSON(u))
}

func (a *App) apiCrearUsuario(w http.ResponseWriter, r *http.Request) {
	var p peticionCrearUsuario
	if err := decodificarJSON(r, &p); err != nil {
		escribirError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}
	if p.Password2 == "" {
		p.Password2 = p.Password
	}

	u, err := a.UsuarioSvc.Registrar(r.Context(), services.DatosRegistro{
		Nombre:    p.Nombre,
		Email:     p.Email,
		Password:  p.Password,
		Password2: p.Password2,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailYaExiste):
			escribirError(w, http.StatusConflict, err.Error())
		case esErrorDeValidacion(err):
			escribirError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("ERROR API crear usuario: %v", err)
			escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		}
		return
	}
	w.Header().Set("Location", "/api/usuarios/"+strconv.FormatInt(u.ID, 10))
	escribirJSON(w, http.StatusCreated, aUsuarioJSON(u))
}

// apiActualizarUsuario → PUT /api/usuarios/{id}.
//
// Solo permite cambiar nombre y/o email. La validación y el UPDATE viven
// ahora en services.UsuarioService.ActualizarDatos: handler ↔ service ↔
// repository, sin SQL filtrado en handlers.
func (a *App) apiActualizarUsuario(w http.ResponseWriter, r *http.Request, id int64) {
	var p peticionEditarUsuario
	if err := decodificarJSON(r, &p); err != nil {
		escribirError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}

	// Comprobar que el usuario existe (404 limpio).
	if _, err := a.UsuarioSvc.ObtenerPorID(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			escribirError(w, http.StatusNotFound, "usuario no encontrado")
			return
		}
		log.Printf("ERROR API obtener usuario %d: %v", id, err)
		escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}

	if err := a.UsuarioSvc.ActualizarDatos(r.Context(), id, p.Nombre, p.Email); err != nil {
		switch {
		case errors.Is(err, services.ErrEmailYaExiste):
			escribirError(w, http.StatusConflict, err.Error())
		case errors.Is(err, services.ErrNombreInvalido), errors.Is(err, services.ErrEmailInvalido):
			escribirError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("ERROR API actualizar usuario %d: %v", id, err)
			escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		}
		return
	}

	u, err := a.UsuarioSvc.ObtenerPorID(r.Context(), id)
	if err != nil {
		log.Printf("ERROR API releer usuario %d: %v", id, err)
		escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}
	escribirJSON(w, http.StatusOK, aUsuarioJSON(u))
}

// apiEliminarUsuario → DELETE /api/usuarios/{id}.
func (a *App) apiEliminarUsuario(w http.ResponseWriter, r *http.Request, actual *models.Usuario, id int64) {
	if _, err := a.UsuarioSvc.ObtenerPorID(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			escribirError(w, http.StatusNotFound, "usuario no encontrado")
			return
		}
		log.Printf("ERROR API obtener usuario %d: %v", id, err)
		escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}

	if err := a.UsuarioSvc.Borrar(r.Context(), actual.ID, id); err != nil {
		switch {
		case errors.Is(err, services.ErrBorrarseASiMismo):
			escribirError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, services.ErrUltimoAdmin):
			escribirError(w, http.StatusForbidden, err.Error())
		default:
			log.Printf("ERROR API borrar usuario %d: %v", id, err)
			escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		}
		return
	}
	log.Printf("API: admin %d eliminó al usuario %d", actual.ID, id)
	escribirJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"id":      id,
		"mensaje": "usuario eliminado",
	})
}

// extraerIDRuta saca el id numérico que va al final de la URL,
// rechazando rutas con segmentos extra.
func extraerIDRuta(ruta, prefijo string) (int64, bool) {
	idStr := strings.TrimPrefix(ruta, prefijo)
	if idStr == "" || strings.Contains(idStr, "/") {
		return 0, false
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// decodificarJSON lee el cuerpo como JSON rechazando campos desconocidos.
func decodificarJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("cuerpo vacío")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
