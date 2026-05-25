package feeds

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"skihub/internal/models"
)

// Fetcher obtiene noticias de fuentes RSS/Atom externas.
// Es seguro para uso concurrente.
type Fetcher struct {
	// HTTPClient se puede sustituir en tests. Si es nil se usa uno con timeout.
	HTTPClient *http.Client
	// DefaultMaxItems es el máximo de noticias por fuente cuando la fuente
	// no especifica su propio límite. Por defecto 10.
	DefaultMaxItems int
}

// NewFetcher crea un Fetcher listo para usar.
func NewFetcher() *Fetcher {
	return &Fetcher{
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		DefaultMaxItems: 10,
	}
}

// FetchSource descarga y parsea una fuente. Devuelve las noticias normalizadas
// y un error solo si la fuente falla completamente. Los campos opcionales
// que no existan en el feed tendrán valores de fallback seguros.
func (f *Fetcher) FetchSource(ctx context.Context, src Source) ([]models.Noticia, error) {
	maxItems := src.MaxItems
	if maxItems <= 0 {
		maxItems = f.DefaultMaxItems
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.FeedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("crear request %s: %w", src.FeedURL, err)
	}
	req.Header.Set("User-Agent", "Snowbreak-NewsBot/1.0 (+https://snowbreak.es)")
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml")

	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", src.FeedURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d desde %s", resp.StatusCode, src.FeedURL)
	}

	// Limitamos la lectura a 2 MB para evitar feeds gigantes.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("leer body %s: %w", src.FeedURL, err)
	}

	noticias, err := parseFeed(body, src, maxItems)
	if err != nil {
		return nil, fmt.Errorf("parsear feed %s: %w", src.FeedURL, err)
	}
	return noticias, nil
}

// ─── Estructuras XML (RSS 2.0 y Atom) ───────────────────────────────────────

// rssFeed cubre RSS 2.0 y también Atom básico (para feeds mixtos).
type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	Enclosure   enclosure `xml:"enclosure"`
	// Media RSS (media:thumbnail, media:content)
	MediaContent   []mediaContent   `xml:"http://search.yahoo.com/mrss/ content"`
	MediaThumbnail []mediaThumbnail `xml:"http://search.yahoo.com/mrss/ thumbnail"`
	// Algunos feeds usan <content:encoded>
	ContentEncoded string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	// GUID para deduplicación alternativa
	GUID string `xml:"guid"`
}

type enclosure struct {
	URL  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

type mediaContent struct {
	URL    string `xml:"url,attr"`
	Medium string `xml:"medium,attr"`
}

type mediaThumbnail struct {
	URL string `xml:"url,attr"`
}

// atomFeed cubre Atom 1.0.
type atomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string     `xml:"title"`
	Links     []atomLink `xml:"link"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
	Updated   string     `xml:"updated"`
	Published string     `xml:"published"`
	ID        string     `xml:"id"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

// ─── Parser principal ────────────────────────────────────────────────────────

func parseFeed(data []byte, src Source, maxItems int) ([]models.Noticia, error) {
	// Intentamos RSS 2.0 primero.
	var rss rssFeed
	if err := xml.Unmarshal(data, &rss); err == nil && len(rss.Channel.Items) > 0 {
		return parseRSSItems(rss.Channel.Items, src, maxItems), nil
	}

	// Intentamos Atom.
	var atom atomFeed
	if err := xml.Unmarshal(data, &atom); err == nil && len(atom.Entries) > 0 {
		return parseAtomEntries(atom.Entries, src, maxItems), nil
	}

	return nil, fmt.Errorf("formato de feed no reconocido (ni RSS 2.0 ni Atom)")
}

func parseRSSItems(items []rssItem, src Source, maxItems int) []models.Noticia {
	var noticias []models.Noticia
	for i, item := range items {
		if i >= maxItems {
			break
		}
		titulo := sanitizarTexto(item.Title)
		if titulo == "" {
			continue
		}
		link := strings.TrimSpace(item.Link)
		if link == "" {
			link = strings.TrimSpace(item.GUID)
		}
		if link == "" {
			continue // sin URL no tiene sentido
		}

		extracto := extraerExtracto(item.Description, item.ContentEncoded)
		imagen := extraerImagenRSS(item)
		fecha := parsearFecha(item.PubDate)

		noticias = append(noticias, normalizarNoticia(titulo, extracto, link, imagen, fecha, src))
	}
	return noticias
}

func parseAtomEntries(entries []atomEntry, src Source, maxItems int) []models.Noticia {
	var noticias []models.Noticia
	for i, entry := range entries {
		if i >= maxItems {
			break
		}
		titulo := sanitizarTexto(entry.Title)
		if titulo == "" {
			continue
		}

		// En Atom el enlace canónico tiene rel="alternate" o está vacío.
		link := ""
		for _, l := range entry.Links {
			if l.Rel == "alternate" || l.Rel == "" {
				link = strings.TrimSpace(l.Href)
				break
			}
		}
		if link == "" {
			link = strings.TrimSpace(entry.ID)
		}
		if link == "" {
			continue
		}

		cuerpo := entry.Summary
		if cuerpo == "" {
			cuerpo = entry.Content
		}
		extracto := extraerExtracto(cuerpo, "")
		fechaStr := entry.Published
		if fechaStr == "" {
			fechaStr = entry.Updated
		}
		fecha := parsearFecha(fechaStr)

		noticias = append(noticias, normalizarNoticia(titulo, extracto, link, "", fecha, src))
	}
	return noticias
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// normalizarNoticia construye una Noticia con todos los campos rellenos.
func normalizarNoticia(titulo, extracto, link, imagen string, fecha time.Time, src Source) models.Noticia {
	if imagen == "" {
		imagen = imagenPorDefecto(src.CategoriaClase)
	}
	return models.Noticia{
		Titulo:         titulo,
		Extracto:       extracto,
		Categoria:      src.Category,
		CategoriaClase: src.CategoriaClase,
		Fecha:          fecha,
		Imagen:         imagen,
		SourceName:     src.Name,
		SourceURL:      src.HomeURL,
		OriginalURL:    link,
		Slug:           generarSlug(titulo),
		IsExternal:     true,
	}
}

// imagenPorDefecto devuelve una imagen de nieve genérica según la categoría.
// Usa imágenes de Unsplash con parámetros fijos para no variar entre recargas.
func imagenPorDefecto(clase string) string {
	switch clase {
	case "nevada":
		return "https://images.unsplash.com/photo-1513342774453-5d76a9768b41?q=80&w=400&h=225&auto=format&fit=crop"
	case "consejos":
		return "https://images.unsplash.com/photo-1551698618-1dfe5d97d256?q=80&w=400&h=225&auto=format&fit=crop"
	case "evento":
		return "https://plus.unsplash.com/premium_photo-1664302791901-52c6159eaf78?q=80&w=400&h=225&auto=format&fit=crop"
	case "seguridad":
		return "https://images.unsplash.com/photo-1732692583018-2345548e4e5a?q=80&w=400&h=225&auto=format&fit=crop"
	case "material":
		return "https://images.unsplash.com/photo-1596473536056-91eadf31189e?q=80&w=400&h=225&auto=format&fit=crop"
	default:
		return "https://images.unsplash.com/photo-1486684338211-1a7ced564b0d?q=80&w=400&h=225&auto=format&fit=crop"
	}
}

// extraerExtracto limpia el HTML y devuelve un resumen de máximo 200 caracteres.
func extraerExtracto(descripcion, contenido string) string {
	texto := descripcion
	if texto == "" {
		texto = contenido
	}
	texto = stripHTML(texto)
	texto = html.UnescapeString(texto)
	texto = strings.TrimSpace(texto)
	texto = strings.Join(strings.Fields(texto), " ") // normalizar espacios
	return truncarUTF8(texto, 200)
}

// extraerImagenRSS busca una URL de imagen en los campos habituales de RSS.
func extraerImagenRSS(item rssItem) string {
	// 1. <enclosure> con tipo imagen
	if item.Enclosure.URL != "" && strings.HasPrefix(item.Enclosure.Type, "image/") {
		return item.Enclosure.URL
	}
	// 2. media:thumbnail
	if len(item.MediaThumbnail) > 0 && item.MediaThumbnail[0].URL != "" {
		return item.MediaThumbnail[0].URL
	}
	// 3. media:content medium="image"
	for _, mc := range item.MediaContent {
		if mc.Medium == "image" && mc.URL != "" {
			return mc.URL
		}
	}
	// 4. <img src="..."> dentro del description
	if img := extraerImgSrc(item.Description); img != "" {
		return img
	}
	if img := extraerImgSrc(item.ContentEncoded); img != "" {
		return img
	}
	return ""
}

var reImgSrc = regexp.MustCompile(`(?i)<img[^>]+src=["']([^"']+)["']`)

func extraerImgSrc(html string) string {
	m := reImgSrc.FindStringSubmatch(html)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

var reHTMLTag = regexp.MustCompile(`<[^>]*>`)
var reHTMLComment = regexp.MustCompile(`<!--.*?-->`)

// stripHTML elimina todas las etiquetas HTML de forma segura (sin usar html/template
// ni parsear el árbol DOM, que sería excesivo para un extracto corto).
func stripHTML(s string) string {
	s = reHTMLComment.ReplaceAllString(s, "")
	s = reHTMLTag.ReplaceAllString(s, " ")
	return s
}

// sanitizarTexto decodifica entidades HTML y elimina espacios extra.
func sanitizarTexto(s string) string {
	s = stripHTML(s)
	s = html.UnescapeString(s)
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

// truncarUTF8 corta el string en el rune boundary más cercano sin superar maxBytes.
func truncarUTF8(s string, maxRunes int) string {
	count := 0
	for i := range s {
		if count >= maxRunes {
			return s[:i] + "…"
		}
		count++
		_ = i
	}
	_ = utf8.RuneCountInString // asegura que el import se use
	return s
}

// ─── Parseo de fechas ─────────────────────────────────────────────────────────

// formatos de fecha que aparecen en feeds RSS/Atom reales.
var fechaFormatos = []string{
	time.RFC1123Z,
	time.RFC1123,
	time.RFC3339,
	time.RFC3339Nano,
	"Mon, 02 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04:05 MST",
	"02 Jan 2006 15:04:05 -0700",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
	"02/01/2006",
}

func parsearFecha(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, formato := range fechaFormatos {
		if t, err := time.Parse(formato, s); err == nil {
			return t.UTC()
		}
	}
	log.Printf("[feeds] no se pudo parsear fecha: %q", s)
	return time.Now().UTC()
}

// ─── Slug ─────────────────────────────────────────────────────────────────────

var reSlugInvalido = regexp.MustCompile(`[^a-z0-9-]`)
var reSlugEspacios = regexp.MustCompile(`\s+`)
var reSlugGuiones = regexp.MustCompile(`-+`)

// generarSlug crea un slug URL-friendly a partir de un título.
func generarSlug(titulo string) string {
	s := strings.ToLower(titulo)
	// Reemplazar caracteres acentuados comunes del español.
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u",
		"ñ", "n", "ü", "u", "à", "a", "è", "e", "ï", "i",
		"ç", "c",
	)
	s = replacer.Replace(s)
	s = reSlugEspacios.ReplaceAllString(s, "-")
	s = reSlugInvalido.ReplaceAllString(s, "")
	s = reSlugGuiones.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 100 {
		s = s[:100]
	}
	return s
}
