package utils

import (
	"math"
	"testing"
)

func TestCoordinateBoundaries(t *testing.T) {
	for _, point := range [][2]float64{{0, 0}, {85, 180}, {-85, -180}, {12.97, 77.59}} {
		if !ValidCoordinates(point[0], point[1]) {
			t.Fatalf("rejected valid coordinate %v", point)
		}
	}
	for _, point := range [][2]float64{{90, 0}, {0, 181}, {math.NaN(), 0}, {0, math.Inf(1)}} {
		if ValidCoordinates(point[0], point[1]) {
			t.Fatalf("accepted invalid coordinate %v", point)
		}
	}
}

func TestDistanceAndEmail(t *testing.T) {
	if d := DistanceMeters(0, 0, 0, 0.001); math.Abs(d-111.195) > 0.1 {
		t.Fatalf("distance %f", d)
	}
	if d := DistanceMeters(0, 0, 0, 180); math.IsNaN(d) {
		t.Fatal("antipodal distance is NaN")
	}
	if email, ok := NormalizeEmail("  Person@Example.com "); !ok || email != "person@example.com" {
		t.Fatal("email not normalized")
	}
	for _, value := range []string{"", "invalid", "Person <person@example.com>", "one@example.com,two@example.com"} {
		if _, ok := NormalizeEmail(value); ok {
			t.Fatalf("accepted %q", value)
		}
	}
}
