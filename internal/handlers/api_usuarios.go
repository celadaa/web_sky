// Package handlers — API REST de usuarios para la PEC 3.
//
// Este archivo añade los 5 endpoints CRUD que consume el cliente JS
// (web/static/admin-api.js). Reutiliza la cadena de capas existente
// (handlers → services → repository) sin tocar la lógica previa:
//
//	GET    /api/usuarios       → listar todos los usuarios (JSON)
//	GET    /api/usuarios/{id}  → obtener un usuario por id (JSON)
//	POST   /api/usuarios       → crear usuario nuevo (JSON, 201)
//	PUT    /api/usuarios/{id}  → actualizar nombre/email (JSON, 200)
//	DELETE /api/usuarios/{id}  → eliminar usuario y cascada (JSON, 200)
//
// Todas las respuestas son application/json. Los códigos de estado
// utilizados son: 200, 201, 400, 401, 403, 404, 405, 409, 500.
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
	"skihub/internal/repository"
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

// peticionCrearUsuario es el cuerpo JSON aceptado en POST /api/usuarios.
// Se exige nombre, email y password (la confirmación es opcional: si se
// envía, debe coincidir con password).
type peticionCrearUsuario struct {
	Nombre    string `json:"nombre"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Password2 string `json:"password2,omitempty"`
}

// peticionEditarUsuario es el cuerpo JSON aceptado en PUT /api/usuarios/{id}.
// Solo permite actualizar nombre y/o email; nunca contraseña ni rol.
type peticionEditarUsuario struct {
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

// respuestaError es el formato uniforme de los mensajes de error JSON.
type respuestaError struct {
	Error string `json:"error"`
}

// aUsuarioJSON convierte un *models.Usuario en su versión segura para JSON.
// Centralizado para garantizar que jamás se filtra el hash de la contraseña.
func aUsuarioJSON(u *models.Usuario) usuarioJSON {
	return usuarioJSON{
		ID:            u.ID,
		Nombre:        u.Nombre,
		Email:         u.Email,
		FechaRegistro: u.FechaRegistro.Format("2006-01-02 15:04:05"),
		EsAdmin:       u.EsAdmin,
	}
}

// escribirJSON serializa el valor dado y lo escribe con el código HTTP
// indicado. Si la serialización falla, devuelve 500 con un mensaje fijo.
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

// escribirError envía una respuesta JSON con el formato {"error":"..."}.
// Es la contrapartida de escribirJSON para los caminos de error.
func escribirError(w http.ResponseWriter, codigo int, msg string) {
	escribirJSON(w, codigo, respuestaError{Error: msg})
}

// requerirAdminAPI es la versión "API" de requerirAdmin: en lugar de
// redirigir al login (lo cual rompería un cliente JSON), devuelve un
// código HTTP claro (401 sin sesión, 403 sin rol). Devuelve nil si la
// petición debe abortarse; en ese caso ya se ha escrito la respuesta.
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

// ApiUsuarios maneja la colección /api/usuarios:
//   - GET  → listar todos los usuarios.
//   - POST → crear un usuario nuevo.
//
// Cualquier otro método devuelve 405 con cabecera Allow.
func (a *App) ApiUsuarios(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.apiListarUsuarios(w, r)
	case http.MethodPost:
		// La creación es escritura: requiere admin autenticado.
		if a.requerirAdminAPI(w, r) == nil {
			return
		}
		a.apiCrearUsuario(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		escribirError(w, http.StatusMethodNotAllowed, "método no permitido")
	}
}

// ApiUsuario maneja el recurso /api/usuarios/{id}:
//   - GET    → obtener un usuario.
//   - PUT    → actualizar nombre y/o email.
//   - DELETE → eliminar (y cascada de favoritos/sesiones).
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
		// Modificación: requiere admin.
		if a.requerirAdminAPI(w, r) == nil {
			return
		}
		a.apiActualizarUsuario(w, r, id)
	case http.MethodDelete:
		// Borrado: requiere admin.
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

// apiListarUsuarios → GET /api/usuarios.
// Devuelve un array JSON con todos los usuarios (sin contraseñas).
func (a *App) apiListarUsuarios(w http.ResponseWriter, r *http.Request) {
	lista, err := a.UsuarioSvc.Listar()
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

// apiObtenerUsuario → GET /api/usuarios/{id}.
// 404 si el id no existe.
func (a *App) apiObtenerUsuario(w http.ResponseWriter, r *http.Request, id int64) {
	u, err := a.UsuarioSvc.ObtenerPorID(id)
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

// apiCrearUsuario → POST /api/usuarios.
// Cuerpo JSON: {"nombre","email","password"[,"password2"]}.
// Reutiliza UsuarioService.Registrar (mismas validaciones que el formulario).
// Devuelve 201 + Location con la URL del recurso creado.
func (a *App) apiCrearUsuario(w http.ResponseWriter, r *http.Request) {
	var p peticionCrearUsuario
	if err := decodificarJSON(r, &p); err != nil {
		escribirError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}
	// Si el cliente no manda password2, asumimos confirmación implícita.
	if p.Password2 == "" {
		p.Password2 = p.Password
	}

	u, err := a.UsuarioSvc.Registrar(services.DatosRegistro{
		Nombre:    p.Nombre,
		Email:     p.Email,
		Password:  p.Password,
		Password2: p.Password2,
	})
	if err != nil {
		// Errores de negocio → 400/409 según el caso.
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
// Solo se permite cambiar nombre y/o email. Para no modificar la lógica
// existente de servicios, el SQL UPDATE concreto se delega al repositorio:
// usamos el método Crear/Actualizar más cercano disponible. Como el repo
// actual no exponía un "ActualizarDatos", lo añadimos *aquí* localmente
// reusando el *sql.DB del repo a través de un método auxiliar.
//
// Devuelve 200 con el usuario actualizado, 404 si no existe, 409 si el
// email ya pertenece a otro usuario, 400 si los datos no son válidos.
func (a *App) apiActualizarUsuario(w http.ResponseWriter, r *http.Request, id int64) {
	var p peticionEditarUsuario
	if err := decodificarJSON(r, &p); err != nil {
		escribirError(w, http.StatusBadRequest, "JSON inválido: "+err.Error())
		return
	}
	nombre := strings.TrimSpace(p.Nombre)
	email := strings.ToLower(strings.TrimSpace(p.Email))

	// Validaciones básicas alineadas con UsuarioService.Registrar.
	if n := len([]rune(nombre)); n < 2 || n > 60 {
		escribirError(w, http.StatusBadRequest, services.ErrNombreInvalido.Error())
		return
	}
	if !validarEmailSimple(email) {
		escribirError(w, http.StatusBadRequest, services.ErrEmailInvalido.Error())
		return
	}

	// Comprobar que el usuario existe.
	u, err := a.UsuarioSvc.ObtenerPorID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			escribirError(w, http.StatusNotFound, "usuario no encontrado")
			return
		}
		log.Printf("ERROR API obtener usuario %d: %v", id, err)
		escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}

	// Si el email cambia, validar que no esté ocupado por otro usuario.
	if email != u.Email {
		otro, err := a.UsuarioSvc.Repo.BuscarPorEmail(email)
		if err == nil && otro != nil && otro.ID != id {
			escribirError(w, http.StatusConflict, services.ErrEmailYaExiste.Error())
			return
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("ERROR API buscar email %s: %v", email, err)
			escribirError(w, http.StatusInternalServerError, "error interno del servidor")
			return
		}
	}

	// Actualizar a través del repositorio (método añadido en repository/usuario.go,
	// no toca lógica previa: solo cambia nombre y email).
	if err := actualizarDatosBasicos(a.UsuarioSvc.Repo, id, nombre, email); err != nil {
		if errors.Is(err, repository.ErrEmailYaRegistrado) {
			escribirError(w, http.StatusConflict, services.ErrEmailYaExiste.Error())
			return
		}
		log.Printf("ERROR API actualizar usuario %d: %v", id, err)
		escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}

	// Releer para devolver el estado actualizado.
	u, err = a.UsuarioSvc.ObtenerPorID(id)
	if err != nil {
		log.Printf("ERROR API releer usuario %d: %v", id, err)
		escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}
	escribirJSON(w, http.StatusOK, aUsuarioJSON(u))
}

// apiEliminarUsuario → DELETE /api/usuarios/{id}.
// Reutiliza UsuarioService.Borrar, que aplica las salvaguardas:
//   - no borrarse a uno mismo
//   - no borrar el último administrador
//
// La cascada de favoritos y sesiones la hace SQLite (ON DELETE CASCADE).
func (a *App) apiEliminarUsuario(w http.ResponseWriter, r *http.Request, actual *models.Usuario, id int64) {
	// Comprobar primero que existe (para devolver 404 limpio).
	if _, err := a.UsuarioSvc.ObtenerPorID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			escribirError(w, http.StatusNotFound, "usuario no encontrado")
			return
		}
		log.Printf("ERROR API obtener usuario %d: %v", id, err)
		escribirError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}

	if err := a.UsuarioSvc.Borrar(actual.ID, id); err != nil {
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
// rechazando rutas con segmentos extra como /api/usuarios/1/foo.
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

// decodificarJSON lee el cuerpo de la petición como JSON, rechazando
// campos desconocidos (más estricto = menos errores silenciosos).
func decodificarJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("cuerpo vacío")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// validarEmailSimple es una validación de formato muy ligera, equivalente
// a la que hace services.UsuarioService.Registrar (no replicamos su regexp
// para no exportar internals; basta con cumplir el patrón mínimo).
func validarEmailSimple(email string) bool {
	if email == "" || len(email) > 120 {
		return false
	}
	at := strings.IndexByte(email, '@')
	if at <= 0 || at == len(email)-1 {
		return false
	}
	dot := strings.LastIndexByte(email, '.')
	if dot < at+2 || dot == len(email)-1 {
		return false
	}
	if strings.ContainsAny(email, " \t\r\n") {
		return false
	}
	return true
}

// actualizarDatosBasicos ejecuta el UPDATE de nombre y email en la tabla
// usuarios. Está aquí (en handlers) y no en repository/ porque el enunciado
// pide NO tocar la lógica existente de repositorios; este auxiliar usa el
// *sql.DB del repo de forma puntual y aislada para esta nueva API.
func actualizarDatosBasicos(repo *repository.UsuarioRepo, id int64, nombre, email string) error {
	_, err := repo.BD.Exec(
		`UPDATE usuarios SET nombre = ?, email = ? WHERE id = ?`,
		nombre, email, id,
	)
	if err != nil {
		// Replica la detección de UNIQUE que usa el repo internamente.
		msg := err.Error()
		if strings.Contains(msg, "UNIQUE constraint") ||
			strings.Contains(msg, "constraint failed: usuarios.email") {
			return repository.ErrEmailYaRegistrado
		}
		return err
	}
	return nil
}
