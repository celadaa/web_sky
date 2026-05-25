package feeds

import (
	"context"
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	// DefaultMaxItems es el maximo de noticias por fuente cuando la fuente
	// no especifica su propio limite. Por defecto 10.
	DefaultMaxItems int
	// StaticDir es la ruta al directorio web/static/ del servidor.
	// Si se configura, las imagenes de RSS se descargan y sirven localmente,
	// evitando problemas de hotlink blocking de sitios externos.
	StaticDir string
}

// NewFetcher crea un Fetcher listo para usar.
func NewFetcher(staticDir string) *Fetcher {
	return &Fetcher{
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		DefaultMaxItems: 10,
		StaticDir:       staticDir,
	}
}

// FetchSource descarga y parsea una fuente. Devuelve las noticias normalizadas
// y un error solo si la fuente falla completamente.
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

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("leer body %s: %w", src.FeedURL, err)
	}

	noticias, err := parseFeed(body, src, maxItems)
	if err != nil {
		return nil, fmt.Errorf("parsear feed %s: %w", src.FeedURL, err)
	}

	// Post-proceso: cachear imagenes en disco para servir desde nuestro servidor.
	// Esto elimina la dependencia de URLs externas en el navegador.
	for i := range noticias {
		noticias[i].Imagen = f.cachearImagen(ctx, noticias[i].Imagen, noticias[i].CategoriaClase)
	}

	return noticias, nil
}

// cachearImagen descarga la imagen a web/static/news-img/ y devuelve la ruta local.
// Si la imagen ya esta cacheada, devuelve la ruta sin re-descargar.
// Si la descarga falla, devuelve la imagen de fallback por categoria de Unsplash.
func (f *Fetcher) cachearImagen(ctx context.Context, imgURL, categoriaClase string) string {
	fallback := imagenPorDefecto(categoriaClase)

	// Sin directorio estatico configurado: usamos Unsplash directamente.
	if f.StaticDir == "" {
		if imgURL == "" {
			return fallback
		}
		return imgURL
	}

	// Si ya es una ruta local o Unsplash (fallback previo), no hacer nada.
	if imgURL == "" ||
		strings.HasPrefix(imgURL, "/static/") ||
		strings.HasPrefix(imgURL, "https://images.unsplash.com") ||
		strings.HasPrefix(imgURL, "https://plus.unsplash.com") {
		if imgURL == "" {
			return fallback
		}
		return imgURL
	}

	// Preparar directorio de cache.
	newsImgDir := filepath.Join(f.StaticDir, "news-img")
	if err := os.MkdirAll(newsImgDir, 0755); err != nil {
		log.Printf("[feeds] no se pudo crear news-img/: %v", err)
		return fallback
	}

	// Nombre de archivo basado en hash MD5 de la URL (extension provisional).
	h := md5.Sum([]byte(imgURL))
	ext := extensionDesdeURL(imgURL)
	filename := fmt.Sprintf("%x%s", h, ext)
	localPath := filepath.Join(newsImgDir, filename)

	// Ya esta en disco: devolver ruta directamente.
	if _, err := os.Stat(localPath); err == nil {
		return "/static/news-img/" + filename
	}

	// Descargar con timeout corto para no bloquear la sincronizacion.
	imgCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	imgReq, err := http.NewRequestWithContext(imgCtx, http.MethodGet, imgURL, nil)
	if err != nil {
		return fallback
	}
	// User-Agent de navegador para saltarse algunos bloqueos de hotlink.
	imgReq.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	// Sin Referer: evita la proteccion de hotlink basada en origen.
	imgReq.Header.Del("Referer")

	imgResp, err := f.HTTPClient.Do(imgReq)
	if err != nil {
		log.Printf("[feeds] imagen no accesible %s: %v", imgURL, err)
		return fallback
	}
	defer imgResp.Body.Close()

	if imgResp.StatusCode != http.StatusOK {
		log.Printf("[feeds] imagen devolvio HTTP %d: %s", imgResp.StatusCode, imgURL)
		return fallback
	}

	// Verificar que es realmente una imagen.
	ct := imgResp.Header.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "image/") {
		log.Printf("[feeds] respuesta no es imagen (Content-Type: %s): %s", ct, imgURL)
		return fallback
	}

	// Ajustar extension segun Content-Type real.
	ext = extensionDesdeContentType(ct, ext)
	filename = fmt.Sprintf("%x%s", h, ext)
	localPath = filepath.Join(newsImgDir, filename)

	data, err := io.ReadAll(io.LimitReader(imgResp.Body, 5<<20))
	if err != nil || len(data) < 100 {
		return fallback
	}

	if err := os.WriteFile(localPath, data, 0644); err != nil {
		log.Printf("[feeds] error guardando imagen en disco: %v", err)
		return fallback
	}

	log.Printf("[feeds] imagen cacheada: %s -> /static/news-img/%s", imgURL, filename)
	return "/static/news-img/" + filename
}

func extensionDesdeURL(u string) string {
	u = strings.ToLower(strings.Split(u, "?")[0])
	switch {
	case strings.HasSuffix(u, ".png"):
		return ".png"
	case strings.HasSuffix(u, ".webp"):
		return ".webp"
	case strings.HasSuffix(u, ".gif"):
		return ".gif"
	default:
		return ".jpg"
	}
}

func extensionDesdeContentType(ct, defecto string) string {
	switch {
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "webp"):
		return ".webp"
	case strings.Contains(ct, "gif"):
		return ".gif"
	case strings.Contains(ct, "jpeg"), strings.Contains(ct, "jpg"):
		return ".jpg"
	default:
		return defecto
	}
}

// ─── Estructuras XML (RSS 2.0 y Atom) ───────────────────────────────────────

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
	// GUID para deduplicacion alternativa
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
	var rss rssFeed
	if err := xml.Unmarshal(data, &rss); err == nil && len(rss.Channel.Items) > 0 {
		return parseRSSItems(rss.Channel.Items, src, maxItems), nil
	}

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
			continue
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

func imagenPorDefecto(clase string) string {
	switch clase {
	case "nevada":
		return "https://images.unsplash.com/photo-1513342774453-5d76a9768b41?q=80&w=800&auto=format&fit=crop"
	case "consejos":
		return "https://images.unsplash.com/photo-1551698618-1dfe5d97d256?q=80&w=800&auto=format&fit=crop"
	case "evento":
		return "https://plus.unsplash.com/premium_photo-1664302791901-52c6159eaf78?q=80&w=800&auto=format&fit=crop"
	case "seguridad":
		return "https://images.unsplash.com/photo-1732692583018-2345548e4e5a?q=80&w=800&auto=format&fit=crop"
	case "material":
		return "https://images.unsplash.com/photo-1596473536056-91eadf31189e?q=80&w=800&auto=format&fit=crop"
	default:
		return "https://images.unsplash.com/photo-1486684338211-1a7ced564b0d?q=80&w=800&auto=format&fit=crop"
	}
}

func extraerExtracto(descripcion, contenido string) string {
	texto := descripcion
	if texto == "" {
		texto = contenido
	}
	texto = stripHTML(texto)
	texto = html.UnescapeString(texto)
	texto = strings.TrimSpace(texto)
	texto = strings.Join(strings.Fields(texto), " ")
	return truncarUTF8(texto, 200)
}

func extraerImagenRSS(item rssItem) string {
	if item.Enclosure.URL != "" && strings.HasPrefix(item.Enclosure.Type, "image/") {
		return item.Enclosure.URL
	}
	if len(item.MediaThumbnail) > 0 && item.MediaThumbnail[0].URL != "" {
		return item.MediaThumbnail[0].URL
	}
	for _, mc := range item.MediaContent {
		if mc.Medium == "image" && mc.URL != "" {
			return mc.URL
		}
	}
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

func stripHTML(s string) string {
	s = reHTMLComment.ReplaceAllString(s, "")
	s = reHTMLTag.ReplaceAllString(s, " ")
	return s
}

func sanitizarTexto(s string) string {
	s = stripHTML(s)
	s = html.UnescapeString(s)
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

func truncarUTF8(s string, maxRunes int) string {
	count := 0
	for i := range s {
		if count >= maxRunes {
			return s[:i] + "..."
		}
		count++
		_ = i
	}
	_ = utf8.RuneCountInString
	return s
}

// ─── Parseo de fechas ─────────────────────────────────────────────────────────

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

func generarSlug(titulo string) string {
	s := strings.ToLower(titulo)
	replacer := strings.NewReplacer(
		"a", "a", "e", "e", "i", "i", "o", "o", "u", "u",
		"n", "n", "u", "u", "a", "a", "e", "e", "i", "i",
		"c", "c",
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
