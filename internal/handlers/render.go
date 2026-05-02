// Package handlers contiene los controladores HTTP de SnowBreak.
// Cada archivo agrupa los manejadores de una sección de la web.
package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// Cache carga y cachea todas las plantillas al arrancar el servidor.
// Se genera un *template.Template por página (index, estaciones, etc.)
// enlazando cada una con el layout base.
type Cache map[string]*template.Template

// CargarPlantillas busca todas las .tmpl en el directorio dado y prepara
// un template por página que incluye layout.tmpl (con navbar y footer).
func CargarPlantillas(dir string) (Cache, error) {
	cache := Cache{}
	paginas := []string{
		"index", "estaciones", "estacion", "noticias",
		"registro", "registro_ok", "error",
		"login", "favoritos", "cambiar_password",
		"admin_usuarios", "admin_usuario", "legal",
		"forfaits", "cesta", "pistas",
	}
	layout := filepath.Join(dir, "layout.tmpl")
	for _, p := range paginas {
		ts, err := template.ParseFiles(layout, filepath.Join(dir, p+".tmpl"))
		if err != nil {
			return nil, err
		}
		cache[p] = ts
	}
	log.Printf("Cargadas %d plantillas desde %s", len(cache), dir)
	return cache, nil
}

// render ejecuta la plantilla en un bytes.Buffer antes de escribir la
// respuesta, de modo que si la renderización falla podamos devolver un
// 500 limpio sin enviar HTML parcial.
func render(w http.ResponseWriter, r *http.Request, c Cache, pagina string, datos any) {
	ts, ok := c[pagina]
	if !ok {
		http.Error(w, "plantilla no encontrada: "+pagina, http.StatusInternalServerError)
		log.Printf("ERROR plantilla inexistente: %s", pagina)
		return
	}
	var buf bytes.Buffer
	if err := ts.ExecuteTemplate(&buf, "layout", datos); err != nil {
		log.Printf("ERROR renderizando %s: %v", pagina, err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("ERROR escribiendo respuesta: %v", err)
	}
}
