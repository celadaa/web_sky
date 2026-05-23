package services

import (
	"strings"
	"testing"
)

// TestHashTokenDeterministicoYIrreversible
//   - el hash del mismo token plano siempre da el mismo resultado.
//   - el hash NO es el propio token (no se filtraría leyendo BD).
//   - el hash es hex SHA-256 → 64 chars hex.
func TestHashTokenDeterministicoYIrreversible(t *testing.T) {
	plano := "token-de-prueba"
	h1 := HashToken(plano)
	h2 := HashToken(plano)
	if h1 != h2 {
		t.Fatalf("HashToken no es determinista: %q vs %q", h1, h2)
	}
	if h1 == plano {
		t.Fatal("HashToken devolvió el plano sin hashear")
	}
	if len(h1) != 64 {
		t.Fatalf("hash debería ser de 64 chars hex, got %d (%q)", len(h1), h1)
	}
	for _, c := range h1 {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("hash con caracteres fuera de hex: %q", h1)
		}
	}
}

// TestGenerarTokenVerificacionUnicoYConsistente
//   - cada llamada devuelve un plano distinto (alta entropía).
//   - el hash devuelto coincide con HashToken(plano).
//   - el plano es base64 URL-safe sin padding y de longitud esperada.
func TestGenerarTokenVerificacionUnicoYConsistente(t *testing.T) {
	plano1, hash1, err := GenerarTokenVerificacion()
	if err != nil {
		t.Fatalf("GenerarTokenVerificacion error: %v", err)
	}
	plano2, hash2, err := GenerarTokenVerificacion()
	if err != nil {
		t.Fatalf("GenerarTokenVerificacion error: %v", err)
	}
	if plano1 == plano2 {
		t.Fatal("dos tokens generados consecutivos son iguales (entropía cero?)")
	}
	if HashToken(plano1) != hash1 {
		t.Fatal("hash devuelto no coincide con HashToken(plano)")
	}
	if HashToken(plano2) != hash2 {
		t.Fatal("hash devuelto no coincide con HashToken(plano) (2)")
	}
	// 32 bytes en base64 raw URL = 43 chars (sin '=' de padding).
	if len(plano1) != 43 {
		t.Fatalf("plano debería ser de 43 chars en base64 raw url, got %d", len(plano1))
	}
	// No debería contener '+' '/' '=' (esos no son URL-safe).
	if strings.ContainsAny(plano1, "+/=") {
		t.Fatalf("plano contiene caracteres no URL-safe: %q", plano1)
	}
}

// TestGenerarEstadoOAuth32Bytes
//   - state es base64 URL-safe sin padding.
//   - dos states consecutivos son distintos.
func TestGenerarEstadoOAuth32Bytes(t *testing.T) {
	s1, err := GenerarEstado()
	if err != nil {
		t.Fatalf("GenerarEstado: %v", err)
	}
	s2, err := GenerarEstado()
	if err != nil {
		t.Fatalf("GenerarEstado: %v", err)
	}
	if s1 == s2 {
		t.Fatal("dos states consecutivos son iguales")
	}
	if len(s1) != 43 {
		t.Fatalf("state debería ser 43 chars, got %d (%q)", len(s1), s1)
	}
	if strings.ContainsAny(s1, "+/=") {
		t.Fatalf("state contiene caracteres no URL-safe: %q", s1)
	}
}
