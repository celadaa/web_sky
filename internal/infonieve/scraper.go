package infonieve

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Estado normalizado de apertura de una estación o pista.
type Estado string

const (
	EstadoAbierta Estado = "abierta"
	EstadoCerrada Estado = "cerrada"
	EstadoParcial Estado = "parcial"
)

// Fraccion representa "abiertos/total" en remontes, pistas o km.
// Se usa null cuando el dato no aparece en la fuente.
type Fraccion struct {
	Abiertos *float64 `json:"abiertos"`
	Total    *float64 `json:"total"`
}

// Estacion es la representación resumida que aparece en los listados.
type Estacion struct {
	Nombre       string   `json:"nombre"`
	Slug         string   `json:"slug"`
	URL          string   `json:"url"`
	Estado       Estado   `json:"estado"`
	Remontes     Fraccion `json:"remontes"`
	Pistas       Fraccion `json:"pistas"`
	Kilometros   Fraccion `json:"kilometros"`
	NieveCm      *float64 `json:"nieve_cm"`
	CalidadNieve string   `json:"calidad_nieve,omitempty"`
	Temperatura  string   `json:"temperatura,omitempty"`
}

// PistaDetalle es una pista individual dentro de la ficha de una estación.
type PistaDetalle struct {
	Nombre     string   `json:"nombre"`
	Dificultad string   `json:"dificultad,omitempty"`
	Estado     Estado   `json:"estado,omitempty"`
	LongitudM  *float64 `json:"longitud_m,omitempty"`
}

// RemonteDetalle es un remonte individual.
type RemonteDetalle struct {
	Nombre string `json:"nombre"`
	Tipo   string `json:"tipo,omitempty"`
	Estado Estado `json:"estado,omitempty"`
}

// Webcam expone una imagen pública de la estación.
type Webcam struct {
	URL         string `json:"url"`
	Descripcion string `json:"descripcion,omitempty"`
}

// EstacionDetalle es la ficha completa con pistas, remontes y webcams.
type EstacionDetalle struct {
	Slug             string           `json:"slug"`
	Nombre           string           `json:"nombre"`
	Estado           Estado           `json:"estado"`
	Actualizado      time.Time        `json:"actualizado"`
	NieveEspesorCm   *float64         `json:"nieve_cm"`
	NieveCalidad     string           `json:"calidad_nieve,omitempty"`
	PistasResumen    Fraccion         `json:"pistas_resumen"`
	Kilometros       Fraccion         `json:"kilometros"`
	RemontesResumen  Fraccion         `json:"remontes_resumen"`
	Temperatura      string           `json:"temperatura,omitempty"`
	Pistas           []PistaDetalle   `json:"pistas"`
	Remontes         []RemonteDetalle `json:"remontes"`
	Webcams          []Webcam         `json:"webcams"`
}

// ─── Helpers de parseo ────────────────────────────────────────────────────────

var (
	reNumero    = regexp.MustCompile(`[^\d.\-]`)
	reCm        = regexp.MustCompile(`(?i)(\d+(?:[,.]\d+)?)\s*CM`)
	reSlugURL   = regexp.MustCompile(`estacion-esqui/([^/]+)/`)
	rePartePref = regexp.MustCompile(`(?i)(Parte del|Actualizado el|Actualizado)\s.+$`)
)

func parseEstado(text string) Estado {
	t := strings.ToLower(strings.TrimSpace(text))
	switch {
	case strings.HasPrefix(t, "abierta") || t == "a":
		return EstadoAbierta
	case strings.HasPrefix(t, "cerrada") || t == "c":
		return EstadoCerrada
	case strings.Contains(t, "parcial"):
		return EstadoParcial
	}
	return ""
}

// parseNum extrae el primer número de la cadena (admite coma decimal).
// Devuelve nil si no encuentra ninguno.
func parseNum(text string) *float64 {
	if text == "" {
		return nil
	}
	clean := strings.ReplaceAll(text, ",", ".")
	clean = reNumero.ReplaceAllString(clean, "")
	if clean == "" || clean == "-" || clean == "." {
		return nil
	}
	n, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return nil
	}
	return &n
}

// splitFraction separa "4/34" o "-/16KM" en {abiertos, total}.
func splitFraction(text string) Fraccion {
	clean := strings.ReplaceAll(strings.TrimSpace(text), "KM", "")
	clean = strings.ReplaceAll(clean, "km", "")
	clean = strings.TrimSpace(clean)
	if !strings.Contains(clean, "/") {
		return Fraccion{}
	}
	partes := strings.SplitN(clean, "/", 2)
	return Fraccion{
		Abiertos: parseNum(partes[0]),
		Total:    parseNum(partes[1]),
	}
}

// parseNieve obtiene el espesor en cm y la calidad del texto agregado
// que devuelve infonieve (ej. "120 CMDura/Primavera").
func parseNieve(text string) (*float64, string) {
	clean := strings.TrimSpace(text)
	var nieveCm *float64
	if m := reCm.FindStringSubmatch(clean); len(m) > 1 {
		nieveCm = parseNum(m[1])
	}
	calidad := reCm.ReplaceAllString(clean, "")
	calidad = strings.Map(func(r rune) rune {
		if r == '-' || r == '/' { // las barras separan "Dura/Primavera"
			return ' '
		}
		return r
	}, calidad)
	calidad = strings.Join(strings.Fields(calidad), " ")
	if calidad == "" || strings.EqualFold(calidad, "CM") {
		return nieveCm, ""
	}
	return nieveCm, calidad
}

func parseMeteo(text string) string {
	t := strings.TrimSpace(text)
	if t == "" || t == "-" {
		return ""
	}
	return t
}

// extraerSlug obtiene el slug de la estación a partir de un href.
func extraerSlug(href string) string {
	m := reSlugURL.FindStringSubmatch(href)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

// limpiarNombre elimina sufijos como "Actualizado el ..." o "Parte del ...".
func limpiarNombre(s string) string {
	s = rePartePref.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// ─── Listados ─────────────────────────────────────────────────────────────────

// ListarEstaciones extrae el listado completo de estaciones desde la
// página /estaciones-esqui/. Si esa lista resulta vacía (porque la web
// cambia), cae al parte de nieve general como fallback.
func (c *Client) ListarEstaciones() ([]Estacion, error) {
	body, err := c.fetch(c.base + "/estaciones-esqui/")
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}
	lista := parseListado(doc)
	if len(lista) == 0 {
		// Fallback: parte general
		return c.ParteNieve("")
	}
	return lista, nil
}

// ParteNieve devuelve el parte de nieve. Si region es "", trae el parte
// general; si es un slug válido, el parte de esa región.
func (c *Client) ParteNieve(region string) ([]Estacion, error) {
	urlStr := c.base + "/parte-de-nieve/"
	if region != "" {
		urlStr = c.base + "/parte-de-nieve/" + region + "/"
	}
	body, err := c.fetch(urlStr)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}
	return parseListado(doc), nil
}

func parseListado(doc *goquery.Document) []Estacion {
	var resultado []Estacion
	vistos := map[string]bool{}

	doc.Find(".list_parte .tbody .tr").Each(func(_ int, row *goquery.Selection) {
		cols := row.Children()
		if cols.Length() < 3 {
			return
		}

		col0 := cols.Eq(0)
		nombreRaw := strings.TrimSpace(col0.Find("strong").First().Text())
		if nombreRaw == "" {
			nombreRaw = strings.TrimSpace(col0.Text())
		}
		nombre := limpiarNombre(nombreRaw)
		if nombre == "" || len(nombre) > 100 {
			return
		}

		href, _ := row.Find("a").First().Attr("href")
		slug := extraerSlug(href)
		if slug == "" {
			// Algunas filas no incluyen "estacion-esqui" en el href; usamos
			// el último segmento como aproximación.
			parts := strings.Split(strings.Trim(href, "/"), "/")
			if len(parts) > 0 {
				slug = parts[len(parts)-1]
			}
		}
		if slug == "" || vistos[slug] {
			return
		}
		vistos[slug] = true

		e := Estacion{
			Nombre: nombre,
			Slug:   slug,
			URL:    BaseURL + "/estacion-esqui/" + slug + "/",
		}
		if cols.Length() >= 2 {
			e.Estado = parseEstado(cols.Eq(1).Text())
		}
		// El layout de /estaciones-esqui es ligeramente distinto al
		// /parte-de-nieve. Detectamos heurísticamente por el número de
		// columnas que tenemos.
		switch {
		case cols.Length() >= 7: // /parte-de-nieve: estado | remontes | pistas | km | nieve | meteo
			e.Remontes = splitFraction(cols.Eq(2).Text())
			e.Pistas = splitFraction(cols.Eq(3).Text())
			e.Kilometros = splitFraction(cols.Eq(4).Text())
			cm, calidad := parseNieve(cols.Eq(5).Text())
			e.NieveCm = cm
			e.CalidadNieve = calidad
			e.Temperatura = parseMeteo(cols.Eq(6).Text())
		case cols.Length() >= 6: // mismo layout sin meteo
			e.Remontes = splitFraction(cols.Eq(2).Text())
			e.Pistas = splitFraction(cols.Eq(3).Text())
			e.Kilometros = splitFraction(cols.Eq(4).Text())
			cm, calidad := parseNieve(cols.Eq(5).Text())
			e.NieveCm = cm
			e.CalidadNieve = calidad
		case cols.Length() >= 3: // /estaciones-esqui: estado | km | pistas | nieve | meteo
			e.Kilometros = splitFraction(cols.Eq(2).Text())
			if cols.Length() > 3 {
				e.Pistas = splitFraction(cols.Eq(3).Text())
			}
			if cols.Length() > 4 {
				cm, calidad := parseNieve(cols.Eq(4).Text())
				e.NieveCm = cm
				e.CalidadNieve = calidad
			}
			if cols.Length() > 5 {
				e.Temperatura = parseMeteo(cols.Eq(5).Text())
			}
		}
		resultado = append(resultado, e)
	})

	return resultado
}

// ─── Detalle de estación ─────────────────────────────────────────────────────

// Estacion devuelve la ficha completa (pistas individuales, remontes,
// webcams) a partir del slug.
func (c *Client) Estacion(slug string) (*EstacionDetalle, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug vacío")
	}
	body, err := c.fetch(c.base + "/estacion-esqui/" + slug + "/parte-de-nieve/")
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	det := &EstacionDetalle{
		Slug:        slug,
		Actualizado: time.Now().UTC(),
		Pistas:      []PistaDetalle{},
		Remontes:    []RemonteDetalle{},
		Webcams:     []Webcam{},
	}

	// Nombre: primero el <h1>, luego el <title>.
	det.Nombre = strings.TrimSpace(doc.Find("h1").First().Text())
	if det.Nombre == "" {
		title := doc.Find("title").Text()
		if i := strings.IndexAny(title, "-|"); i > 0 {
			det.Nombre = strings.TrimSpace(title[:i])
		} else {
			det.Nombre = strings.TrimSpace(title)
		}
	}
	if det.Nombre == "" {
		det.Nombre = slug
	}

	// Resumen (primera fila del .list_parte).
	primera := doc.Find(".list_parte .tbody .tr").First()
	if primera.Length() > 0 {
		cols := primera.Children()
		if cols.Length() >= 6 {
			det.Estado = parseEstado(cols.Eq(1).Text())
			det.RemontesResumen = splitFraction(cols.Eq(2).Text())
			det.PistasResumen = splitFraction(cols.Eq(3).Text())
			det.Kilometros = splitFraction(cols.Eq(4).Text())
			cm, calidad := parseNieve(cols.Eq(5).Text())
			det.NieveEspesorCm = cm
			det.NieveCalidad = calidad
			if cols.Length() > 6 {
				det.Temperatura = parseMeteo(cols.Eq(6).Text())
			}
		}
	}

	// Pistas individuales (varios layouts posibles).
	doc.Find(`[class*="list_pista"] .tr, [id*="pistas"] tr, table.pistas tr`).Each(func(_ int, row *goquery.Selection) {
		cols := row.Find(`td, [class*="td"]`)
		if cols.Length() < 2 {
			return
		}
		nombre := strings.TrimSpace(cols.Eq(0).Text())
		if nombre == "" || len(nombre) > 80 || strings.EqualFold(nombre, "pista") {
			return
		}
		p := PistaDetalle{Nombre: nombre}
		if cols.Length() > 1 {
			p.Dificultad = strings.TrimSpace(cols.Eq(1).Text())
		}
		if cols.Length() > 2 {
			p.Estado = parseEstado(cols.Eq(2).Text())
		}
		if cols.Length() > 3 {
			p.LongitudM = parseNum(cols.Eq(3).Text())
		}
		det.Pistas = append(det.Pistas, p)
	})

	// Remontes.
	doc.Find(`[class*="list_remonte"] .tr, [id*="remontes"] tr, table.remontes tr`).Each(func(_ int, row *goquery.Selection) {
		nombre := strings.TrimSpace(row.Children().First().Text())
		if nombre == "" || len(nombre) > 80 {
			return
		}
		cols := row.Find(`td, [class*="td"]`)
		r := RemonteDetalle{Nombre: nombre}
		if cols.Length() > 1 {
			r.Tipo = strings.TrimSpace(cols.Eq(1).Text())
		}
		if cols.Length() > 2 {
			r.Estado = parseEstado(cols.Eq(2).Text())
		}
		det.Remontes = append(det.Remontes, r)
	})

	// Webcams (hasta 20 únicas por src).
	seen := map[string]bool{}
	doc.Find(`img[src*="webcam"], [class*="webcam"] img, .box_webcam img`).Each(func(_ int, img *goquery.Selection) {
		src, _ := img.Attr("src")
		if src == "" {
			src, _ = img.Attr("data-src")
		}
		if src == "" || seen[src] || len(det.Webcams) >= 20 {
			return
		}
		seen[src] = true
		full := src
		if !strings.HasPrefix(src, "http") {
			full = BaseURL + src
		}
		alt, _ := img.Attr("alt")
		det.Webcams = append(det.Webcams, Webcam{URL: full, Descripcion: alt})
	})

	// Truncamos para no devolver listas absurdas si el HTML cambia.
	if len(det.Pistas) > 100 {
		det.Pistas = det.Pistas[:100]
	}
	if len(det.Remontes) > 60 {
		det.Remontes = det.Remontes[:60]
	}
	return det, nil
}
