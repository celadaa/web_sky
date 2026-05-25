// Package feeds gestiona la obtención de noticias de fuentes externas (RSS/Atom).
//
// Para añadir una nueva fuente basta con añadir una entrada a DefaultSources
// (o crear tu propio slice de Source y pasarlo al Fetcher). No es necesario
// tocar ninguna plantilla ni handler.
package feeds

// Source describe una fuente de noticias externa.
type Source struct {
	// Nombre visible de la fuente (p.ej. "Nevasport").
	Name string
	// URL del feed RSS o Atom.
	FeedURL string
	// URL raíz de la fuente (para el enlace "Ir a la fuente").
	HomeURL string
	// Categoría por defecto para las noticias de esta fuente.
	// Debe coincidir con un valor de category_class: nevada | consejos | evento | general | actualidad | seguridad | material
	Category string
	// CategoriaClase es la clase CSS que se aplica al tag de categoría.
	CategoriaClase string
	// MaxItems limita cuántas noticias se importan por ciclo de esta fuente.
	// 0 = usar el valor global del Fetcher.
	MaxItems int
	// Active indica si esta fuente está habilitada.
	Active bool
}

// DefaultSources es la lista de fuentes activas por defecto.
//
// ╔══════════════════════════════════════════════════════════════════════╗
// ║  AÑADIR NUEVA FUENTE: Agrega una entrada aquí con Active: true.     ║
// ║  El sistema la detectará en el próximo ciclo de sincronización      ║
// ║  (cada 2 horas) sin necesidad de reiniciar el servidor.             ║
// ╚══════════════════════════════════════════════════════════════════════╝
var DefaultSources = []Source{
	{
		Name:           "Nevasport",
		FeedURL:        "https://www.nevasport.com/rss/rss.php",
		HomeURL:        "https://www.nevasport.com",
		Category:       "Actualidad",
		CategoriaClase: "nevada",
		MaxItems:       10,
		Active:         true,
	},
	{
		Name:           "RFEDI – Federación Española Deportes de Invierno",
		FeedURL:        "https://www.rfedi.es/feed/",
		HomeURL:        "https://www.rfedi.es",
		Category:       "Eventos",
		CategoriaClase: "evento",
		MaxItems:       8,
		Active:         true,
	},
	{
		Name:           "Atudem – Estaciones de Esquí",
		FeedURL:        "https://www.atudem.es/feed/",
		HomeURL:        "https://www.atudem.es",
		Category:       "Estaciones",
		CategoriaClase: "general",
		MaxItems:       8,
		Active:         true,
	},
	{
		Name:           "Snow-Forecast",
		FeedURL:        "https://www.snow-forecast.com/rss/news.xml",
		HomeURL:        "https://www.snow-forecast.com",
		Category:       "Nevada",
		CategoriaClase: "nevada",
		MaxItems:       8,
		Active:         true,
	},
	{
		Name:           "Freeskier Magazine",
		FeedURL:        "https://freeskier.com/feed/",
		HomeURL:        "https://freeskier.com",
		Category:       "Material",
		CategoriaClase: "material",
		MaxItems:       6,
		Active:         false, // Activar si se quiere contenido en inglés
	},
}

// claseCSSParaCategoria devuelve una clase CSS válida dado un nombre de categoría.
// Se usa como fallback si la fuente no define CategoriaClase.
func claseCSSParaCategoria(cat string) string {
	switch cat {
	case "Nevada", "Nevadas", "Nieve":
		return "nevada"
	case "Consejos", "Guías", "Técnica":
		return "consejos"
	case "Evento", "Eventos", "Competición", "Competiciones":
		return "evento"
	case "Seguridad", "Aludes", "Rescate":
		return "seguridad"
	case "Material", "Equipamiento":
		return "material"
	case "Estaciones":
		return "general"
	default:
		return "general"
	}
}

// ActiveSources devuelve solo las fuentes con Active: true.
func ActiveSources() []Source {
	var active []Source
	for _, s := range DefaultSources {
		if s.Active {
			if s.CategoriaClase == "" {
				s.CategoriaClase = claseCSSParaCategoria(s.Category)
			}
			active = append(active, s)
		}
	}
	return active
}
