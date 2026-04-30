package services

import (
	"math"
	"testing"
)

func TestDistanciaHaversineKm(t *testing.T) {
	// Madrid (40.4168, -3.7038) ↔ Sierra Nevada (37.0928, -3.3953) ≈ 372 km.
	d := DistanciaHaversineKm(40.4168, -3.7038, 37.0928, -3.3953)
	if math.Abs(d-372) > 30 {
		t.Errorf("Madrid-SN: got=%.1f km, esperado ~372 km", d)
	}

	// Mismo punto debe ser 0.
	if d := DistanciaHaversineKm(0, 0, 0, 0); d != 0 {
		t.Errorf("punto consigo mismo: got=%v", d)
	}

	// Simétrico.
	a := DistanciaHaversineKm(40, -3, 42, 1)
	b := DistanciaHaversineKm(42, 1, 40, -3)
	if math.Abs(a-b) > 0.001 {
		t.Errorf("distancia no simétrica: %v vs %v", a, b)
	}
}

func TestCoordsValidas(t *testing.T) {
	if !CoordsValidas(40, -3) {
		t.Error("Madrid debería ser válida")
	}
	if CoordsValidas(91, 0) {
		t.Error("lat=91 debería ser inválida")
	}
	if CoordsValidas(0, -181) {
		t.Error("lng=-181 debería ser inválida")
	}
	if CoordsValidas(math.NaN(), 0) {
		t.Error("NaN debería ser inválida")
	}
}
