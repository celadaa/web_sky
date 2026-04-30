package infonieve

// Region representa una de las áreas geográficas en las que infonieve.es
// agrupa las estaciones de esquí.
type Region struct {
	Slug   string `json:"slug"`
	Nombre string `json:"nombre"`
	Pais   string `json:"pais"`
}

// Regiones es el listado completo soportado por infonieve.es. El orden
// se respeta de norte a sur dentro de España y luego países vecinos.
// Los slugs están verificados contra las URLs reales del sitio.
var Regiones = []Region{
	{"pirineo-aragones", "Pirineo Aragonés", "España"},
	{"pirineo-catalan", "Pirineo Catalán", "España"},
	{"sistema-central", "Sistema Central", "España"},
	{"cordillera-cantabrica", "Cordillera Cantábrica", "España"},
	{"sierra-nevada", "Sierra Nevada", "España"},
	{"sistema-iberiano", "Sistema Ibérico", "España"},
	{"galicia", "Galicia", "España"},
	{"andorra", "Andorra", "Andorra"},
	{"pirineo-frances", "Pirineo Francés", "Francia"},
	{"neiges-catalanes", "Neiges Catalanes", "Francia"},
	{"portugal", "Portugal", "Portugal"},
}

// RegionPorSlug devuelve la región cuyo slug coincide o nil.
func RegionPorSlug(slug string) *Region {
	for i := range Regiones {
		if Regiones[i].Slug == slug {
			return &Regiones[i]
		}
	}
	return nil
}
