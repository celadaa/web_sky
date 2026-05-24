package infonieve

import "testing"

func TestSlugPorNombre(t *testing.T) {
	cases := []struct{ nombre, slug string }{
		{"Valdesquí",             "valdesqui"},
		{"Valdesqui",             "valdesqui"},
		{"Puerto de Navacerrada", "navacerrada"},
		{"Navacerrada",           "navacerrada"},
		{"La Pinilla",            "la-pinilla"},
		{"Sierra Nevada",         "sierra-nevada"},
		{"Baqueira Beret",        "baqueira-beret"},
		{"Baqueira",              "baqueira-beret"},
		{"Candanchú",             "candanchu"},
		{"Candanchu",             "candanchu"},
		{"Formigal",              "formigal"},
		{"Astún",                 "astun"},
		{"Alto Campoo",           "alto-campoo"},
		{"La Covatilla",          "sierra-de-bejar-la-covatilla"},
	}
	for _, c := range cases {
		got, ok := SlugPorNombre(c.nombre)
		if !ok || got != c.slug {
			t.Errorf("SlugPorNombre(%q) = (%q, %v), want (%q, true)", c.nombre, got, ok, c.slug)
		}
	}
}
