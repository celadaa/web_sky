package models

import (
	"strconv"
	"time"
)

// Noticia representa una entrada del blog de noticias de SkiHub.
// CategoriaClase se usa en la plantilla para aplicar la clase CSS
// correcta (nevada, consejos, evento, general).
type Noticia struct {
	ID             int64
	Titulo         string
	Extracto       string
	Categoria      string
	CategoriaClase string
	Fecha          time.Time
	Imagen         string
}

// FechaISO devuelve la fecha en formato YYYY-MM-DD, útil para el
// atributo datetime de la etiqueta <time>.
func (n Noticia) FechaISO() string {
	return n.Fecha.Format("2006-01-02")
}

// FechaLarga devuelve la fecha en formato legible en español, por ejemplo
// "14 de Marzo, 2026".
func (n Noticia) FechaLarga() string {
	meses := [...]string{
		"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
		"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre",
	}
	return strconv.Itoa(n.Fecha.Day()) + " de " +
		meses[int(n.Fecha.Month())-1] + ", " +
		strconv.Itoa(n.Fecha.Year())
}
