package models

import (
	"fmt"
	"time"
)

// Noticia representa una entrada de noticias de Snowbreak.
// IsExternal=true significa que viene de una fuente RSS externa.
// CategoriaClase: nevada | consejos | evento | general | seguridad | material
type Noticia struct {
	ID             int64
	Titulo         string
	Extracto       string
	Categoria      string
	CategoriaClase string
	Fecha          time.Time
	Imagen         string

	// Campos solo presentes en noticias externas (RSS/API).
	SourceName  string
	SourceURL   string
	OriginalURL string
	Slug        string
	IsExternal  bool
}

// EnlaceNoticia devuelve la URL del boton "Leer mas".
func (n Noticia) EnlaceNoticia() string {
	if n.IsExternal && n.OriginalURL != "" {
		return n.OriginalURL
	}
	return "#"
}

// AbreEnNuevaVentana indica si el enlace debe abrirse en nueva pestana.
func (n Noticia) AbreEnNuevaVentana() bool {
	return n.IsExternal && n.OriginalURL != ""
}

// FechaISO devuelve la fecha en formato YYYY-MM-DD.
func (n Noticia) FechaISO() string {
	return n.Fecha.Format("2006-01-02")
}

// FechaLarga devuelve la fecha legible: "14 de Marzo, 2026".
func (n Noticia) FechaLarga() string {
	meses := [...]string{
		"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
		"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre",
	}
	return fmt.Sprintf("%d de %s, %d",
		n.Fecha.Day(),
		meses[int(n.Fecha.Month())-1],
		n.Fecha.Year(),
	)
}
