package infonieve

import (
	"bytes"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// HTML snapshot fiel a la estructura real de infonieve.es para que los
// tests pasen sin necesidad de red. Es la misma muestra que usa el
// proyecto JS original (scraper.test.js).
const htmlSnapshot = `
<div class="list_parte">
  <div class="tbody">
    <div class="tr">
      <a class="td tdx1" href="estacion-esqui/alp-2500/parte-de-nieve/">
        <span class="fuentegrande"><strong>Alp 2500</strong></span>
        <span class="fuentemicro">Miércoles 29 de Abril</span>
      </a>
      <a class="td tdx2">AbiertaA</a>
      <a class="td tdx3">4/34</a>
      <a class="td tdx4">9/154</a>
      <a class="td tdx5">17/145KM</a>
      <a class="td tdx6">120 CMDura/Primavera</a>
      <a class="td tdx7">-</a>
      <a class="td tdxgo"></a>
    </div>
    <div class="tr">
      <a class="td tdx1" href="estacion-esqui/la-molina/parte-de-nieve/">
        <span class="fuentegrande"><strong>La Molina</strong></span>
        <span class="fuentemicro">Miércoles 29 de Abril</span>
      </a>
      <a class="td tdx2">CerradaC</a>
      <a class="td tdx3">-/16</a>
      <a class="td tdx4">-/66</a>
      <a class="td tdx5">-/71KM</a>
      <a class="td tdx6">- CM</a>
      <a class="td tdx7">0º/13º</a>
      <a class="td tdxgo"></a>
    </div>
  </div>
</div>`

func TestParseEstado(t *testing.T) {
	cases := []struct {
		in  string
		out Estado
	}{
		{"AbiertaA", EstadoAbierta},
		{"CerradaC", EstadoCerrada},
		{"Apertura parcial", EstadoParcial},
		{"", ""},
		{"Algo raro", ""},
	}
	for _, c := range cases {
		if got := parseEstado(c.in); got != c.out {
			t.Errorf("parseEstado(%q) = %q, esperado %q", c.in, got, c.out)
		}
	}
}

func TestParseNum(t *testing.T) {
	cases := []struct {
		in  string
		out *float64
	}{
		{"120 CM", ptr(120)},
		{"27,65KM", ptr(27.65)},
		{"-", nil},
		{"", nil},
	}
	for _, c := range cases {
		got := parseNum(c.in)
		if (got == nil) != (c.out == nil) {
			t.Errorf("parseNum(%q) nil mismatch: got=%v want=%v", c.in, got, c.out)
			continue
		}
		if got != nil && *got != *c.out {
			t.Errorf("parseNum(%q) = %v, esperado %v", c.in, *got, *c.out)
		}
	}
}

func TestSplitFraction(t *testing.T) {
	cases := []struct {
		in   string
		a, t any
	}{
		{"4/34", 4.0, 34.0},
		{"-/71KM", nil, 71.0},
		{"17/145KM", 17.0, 145.0},
		{"sin barra", nil, nil},
	}
	for _, c := range cases {
		got := splitFraction(c.in)
		comparar(t, "abiertos "+c.in, got.Abiertos, c.a)
		comparar(t, "total "+c.in, got.Total, c.t)
	}
}

func TestParseNieve(t *testing.T) {
	cm, calidad := parseNieve("120 CMDura/Primavera")
	if cm == nil || *cm != 120 {
		t.Errorf("nieve_cm: got=%v want=120", cm)
	}
	if !strings.Contains(strings.ToLower(calidad), "dura") {
		t.Errorf("calidad no contiene 'dura': %q", calidad)
	}

	cm, calidad = parseNieve("- CM")
	if cm != nil || calidad != "" {
		t.Errorf("vacío: cm=%v calidad=%q", cm, calidad)
	}
}

func TestParseListadoSnapshot(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(htmlSnapshot)))
	if err != nil {
		t.Fatalf("parse HTML: %v", err)
	}
	res := parseListado(doc)
	if len(res) != 2 {
		t.Fatalf("esperaba 2 estaciones, obtuve %d", len(res))
	}

	alp := res[0]
	if alp.Nombre != "Alp 2500" {
		t.Errorf("nombre: %q", alp.Nombre)
	}
	if alp.Slug != "alp-2500" {
		t.Errorf("slug: %q", alp.Slug)
	}
	if alp.Estado != EstadoAbierta {
		t.Errorf("estado: %q", alp.Estado)
	}
	if alp.Remontes.Abiertos == nil || *alp.Remontes.Abiertos != 4 {
		t.Errorf("remontes abiertos: %v", alp.Remontes.Abiertos)
	}
	if alp.Pistas.Total == nil || *alp.Pistas.Total != 154 {
		t.Errorf("pistas total: %v", alp.Pistas.Total)
	}
	if alp.Kilometros.Abiertos == nil || *alp.Kilometros.Abiertos != 17 {
		t.Errorf("km abiertos: %v", alp.Kilometros.Abiertos)
	}
	if alp.NieveCm == nil || *alp.NieveCm != 120 {
		t.Errorf("nieve cm: %v", alp.NieveCm)
	}

	mol := res[1]
	if mol.Estado != EstadoCerrada {
		t.Errorf("la molina estado: %q", mol.Estado)
	}
	if mol.Temperatura == "" {
		t.Errorf("temperatura no debería ser vacía")
	}
}

func TestCoordPorSlug(t *testing.T) {
	c, ok := CoordPorSlug("baqueira-beret")
	if !ok || c.Lat < 42 || c.Lat > 43 {
		t.Errorf("baqueira: ok=%v c=%+v", ok, c)
	}
	if _, ok := CoordPorSlug("estacion-imaginaria-xyz"); ok {
		t.Errorf("debería devolver false para slug desconocido")
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func ptr(v float64) *float64 { return &v }

func comparar(t *testing.T, etiqueta string, got *float64, want any) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Errorf("%s: got=%v want=nil", etiqueta, *got)
		}
		return
	}
	w := want.(float64)
	if got == nil || *got != w {
		t.Errorf("%s: got=%v want=%v", etiqueta, got, w)
	}
}
