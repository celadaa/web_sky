package infonieve

import "strings"

// Coord es una latitud/longitud en grados decimales (WGS84).
type Coord struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// coordsBySlug mapea el slug que usa infonieve.es a la posición real
// de la estación. La fuente de los puntos son OpenStreetMap y las webs
// oficiales de cada estación; pequeñas desviaciones (< 1 km) son
// aceptables porque el dato sólo se usa para ordenar por proximidad.
var coordsBySlug = map[string]Coord{
	// — Pirineo aragonés —
	"astun":           {42.8125, -0.5253},
	"candanchu":       {42.7833, -0.5167},
	"formigal":        {42.7811, -0.4108},
	"panticosa":       {42.7167, -0.2917},
	"cerler":          {42.5400, 0.4225},
	// — Pirineo catalán —
	"alp-2500":        {42.3506, 1.9408},
	"baqueira-beret":  {42.7000, 0.9333},
	"boi-taull":       {42.5256, 0.8667},
	"espot":           {42.5750, 1.1100},
	"la-molina":       {42.3406, 1.9469},
	"masella":         {42.3675, 1.9347},
	"port-aine":       {42.6219, 1.2358},
	"port-del-comte":  {42.1942, 1.5292},
	"tavascan":        {42.6717, 1.2347},
	"vall-de-nuria":   {42.4006, 2.1531},
	"vallter-2000":    {42.4322, 2.2725},
	// — Sistema Central —
	"la-pinilla":      {41.2917, -3.4583},
	"navacerrada":     {40.7842, -4.0094},
	"valdesqui":       {40.7867, -3.9000},
	"sierra-de-bejar-la-covatilla": {40.3450, -5.7233},
	// — Cordillera Cantábrica —
	"alto-campoo":     {43.0686, -4.4400},
	"fuentes-de-invierno": {43.0700, -5.4500},
	"leitariegos":     {43.0058, -6.4242},
	"san-isidro":      {43.0467, -5.4117},
	"valgrande-pajares": {43.0450, -5.7567},
	// — Sierra Nevada —
	"sierra-nevada":   {37.0928, -3.3953},
	// — Sistema Ibérico —
	"javalambre":      {40.0944, -1.0125},
	"valdelinares":    {40.4133, -0.6233},
	"valdezcaray":     {42.2275, -3.0344},
	// — Galicia —
	"manzaneda":       {42.2675, -7.2467},
	// — Andorra —
	"grandvalira":     {42.5392, 1.7308},
	"grandvalira-pas-de-la-casa":   {42.5450, 1.7378},
	"grandvalira-soldeu-el-tarter": {42.5778, 1.6611},
	"grandvalira-ordino-arcalis":   {42.6342, 1.5158},
	"vallnord":        {42.5786, 1.4917},
	"pal-arinsal":     {42.5786, 1.4917},
	// — Pirineo Francés —
	"saint-lary":      {42.7842, 0.3242},
	"font-romeu":      {42.5158, 2.0367},
	"piau-engaly":     {42.7717, 0.1611},
	"peyragudes":      {42.8014, 0.4500},
	// — Neiges Catalanes —
	"les-angles":      {42.5681, 2.0689},
	"porte-puymorens": {42.5500, 1.7967},
	"pyrenees-2000":   {42.5000, 2.1000},
	"font-romeu-pyrenees-2000": {42.5158, 2.0367},
	// — Portugal —
	"serra-da-estrela": {40.3217, -7.6131},
}

// CoordPorSlug busca por slug y devuelve el punto. Si no se encuentra
// devuelve (Coord{}, false). Tolera variaciones tipográficas.
func CoordPorSlug(slug string) (Coord, bool) {
	if slug == "" {
		return Coord{}, false
	}
	if c, ok := coordsBySlug[slug]; ok {
		return c, true
	}
	// Búsqueda permisiva: si el slug pedido contiene a una clave conocida.
	low := strings.ToLower(slug)
	for k, v := range coordsBySlug {
		if strings.Contains(low, k) || strings.Contains(k, low) {
			return v, true
		}
	}
	return Coord{}, false
}
