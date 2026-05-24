// Package infonieve - tabla unica de mapping nombre->slug.
//
// Esta es la UNICA fuente de verdad para relacionar los nombres de estacion
// que usa la base de datos local con los slugs que usa infonieve.es.
// Ninguna otra capa de la aplicacion debe duplicar esta logica.
package infonieve

import "strings"

// nombreASlug mapea el nombre normalizado de estacion (minusculas, sin tildes,
// espacios colapsados) al slug exacto que usa infonieve.es.
// Incluye variantes habituales de un mismo nombre.
var nombreASlug = map[string]string{
	// Pirineo aragones
	"astun":     "astun",
	"candanchu": "candanchu",
	"formigal":  "formigal",
	"panticosa": "panticosa",
	"cerler":    "cerler",

	// Pirineo catalan
	"baqueira beret": "baqueira-beret",
	"baqueira":       "baqueira-beret",
	"boi taull":      "boi-taull",
	"espot":          "espot",
	"la molina":      "la-molina",
	"masella":        "masella",
	"port aine":      "port-aine",
	"port del comte": "port-del-comte",
	"tavascan":       "tavascan",
	"vall de nuria":  "vall-de-nuria",
	"vallter 2000":   "vallter-2000",
	"alp 2500":       "alp-2500",

	// Sistema Central
	"la pinilla":                   "la-pinilla",
	"navacerrada":                  "navacerrada",
	"puerto de navacerrada":        "navacerrada",
	"valdesqui":                    "valdesqui",
	"sierra de bejar la covatilla": "sierra-de-bejar-la-covatilla",
	"sierra de bejar":              "sierra-de-bejar-la-covatilla",
	"la covatilla":                 "sierra-de-bejar-la-covatilla",

	// Cordillera Cantabrica
	"alto campoo":         "alto-campoo",
	"fuentes de invierno": "fuentes-de-invierno",
	"leitariegos":         "leitariegos",
	"san isidro":          "san-isidro",
	"valgrande pajares":   "valgrande-pajares",

	// Sierra Nevada
	"sierra nevada": "sierra-nevada",

	// Sistema Iberico
	"javalambre":   "javalambre",
	"valdelinares": "valdelinares",
	"valdezcaray":  "valdezcaray",

	// Galicia
	"manzaneda": "manzaneda",

	// Andorra
	"grandvalira": "grandvalira",
	"pal arinsal": "pal-arinsal",
	"vallnord":    "vallnord",

	// Pirineo frances
	"saint lary":               "saint-lary",
	"font romeu":               "font-romeu",
	"font romeu pyrenees 2000": "font-romeu-pyrenees-2000",
	"piau engaly":              "piau-engaly",
	"peyragudes":               "peyragudes",

	// Neiges Catalanes
	"les angles":      "les-angles",
	"porte puymorens": "porte-puymorens",
	"pyrenees 2000":   "pyrenees-2000",

	// Portugal
	"serra da estrela": "serra-da-estrela",
}

// quitarTildes reemplaza los caracteres acentuados del espanol, catalan y
// frances por sus equivalentes ASCII. No depende de golang.org/x/text.
func quitarTildes(s string) string {
	r := strings.NewReplacer(
		"á", "a",
		"à", "a",
		"â", "a",
		"ä", "a",
		"é", "e",
		"è", "e",
		"ê", "e",
		"ë", "e",
		"í", "i",
		"ì", "i",
		"î", "i",
		"ï", "i",
		"ó", "o",
		"ò", "o",
		"ô", "o",
		"ö", "o",
		"ú", "u",
		"ù", "u",
		"û", "u",
		"ü", "u",
		"ñ", "n",
		"ç", "c",
		"Á", "a",
		"À", "a",
		"Â", "a",
		"Ä", "a",
		"É", "e",
		"È", "e",
		"Ê", "e",
		"Ë", "e",
		"Í", "i",
		"Ì", "i",
		"Î", "i",
		"Ï", "i",
		"Ó", "o",
		"Ò", "o",
		"Ô", "o",
		"Ö", "o",
		"Ú", "u",
		"Ù", "u",
		"Û", "u",
		"Ü", "u",
		"Ñ", "n",
		"Ç", "c",
	)
	return r.Replace(s)
}

// NormalizarNombre convierte un nombre de estacion a su forma canonica:
// minusculas, sin tildes, espacios colapsados.
// Es la funcion de normalizacion estandar de la aplicacion.
func NormalizarNombre(nombre string) string {
	s := strings.ToLower(strings.TrimSpace(nombre))
	s = quitarTildes(s)
	return strings.Join(strings.Fields(s), " ")
}

// SlugPorNombre devuelve el slug de infonieve.es correspondiente al nombre
// de estacion dado. Aplica normalizacion antes de buscar.
// Primero busca coincidencia exacta; si no hay, busca por contenido parcial.
// Devuelve ("", false) si no hay ningun mapping conocido.
func SlugPorNombre(nombre string) (string, bool) {
	n := NormalizarNombre(nombre)
	if slug, ok := nombreASlug[n]; ok {
		return slug, true
	}
	// Busqueda permisiva: si la clave esta contenida en el nombre o viceversa.
	for k, v := range nombreASlug {
		if strings.Contains(n, k) || strings.Contains(k, n) {
			return v, true
		}
	}
	return "", false
}
