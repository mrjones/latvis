package location

import (
	"testing"
)

func TestContainsNE(t *testing.T) {
	b := NewBoundingBox(
		Coordinate{Lat: 1.0, Lng: 1.0},
		Coordinate{Lat: 2.0,	Lng: 2.0})
	
	assertTrueM(t, b.Contains(Coordinate{Lat: 1.5, Lng: 1.5}), "In Box")

	assertFalseM(t, b.Contains(Coordinate{Lat: 0.5, Lng: 1.5}), "East")
	assertFalseM(t, b.Contains(Coordinate{Lat: 2.5, Lng: 1.5}), "West")
	assertFalseM(t, b.Contains(Coordinate{Lat: 1.5, Lng: 0.5}), "South")
	assertFalseM(t, b.Contains(Coordinate{Lat: 1.5, Lng: 2.5}), "North")
}

func TestContainsNW(t *testing.T) {
	b := NewBoundingBox(
		Coordinate{Lat: -2.0, Lng: 1.0},
		Coordinate{Lat: -1.0,	Lng: 2.0})
	
	assertTrueM(t, b.Contains(Coordinate{Lat: -1.5, Lng: 1.5}), "In Box")

	assertFalseM(t, b.Contains(Coordinate{Lat: -2.5, Lng: 1.5}), "East")
	assertFalseM(t, b.Contains(Coordinate{Lat: -0.5, Lng: 1.5}), "West")
	assertFalseM(t, b.Contains(Coordinate{Lat: -1.5, Lng: 0.5}), "South")
	assertFalseM(t, b.Contains(Coordinate{Lat: -1.5, Lng: 2.5}), "North")
}

func TestContainsSW(t *testing.T) {
	b := NewBoundingBox(
		Coordinate{Lat: -2.0, Lng: -2.0},
		Coordinate{Lat: -1, Lng: -1.0})
	
	assertTrueM(t, b.Contains(Coordinate{Lat: -1.5, Lng: -1.5}), "In Box")

	assertFalseM(t, b.Contains(Coordinate{Lat: -2.5, Lng: -1.5}), "East")
	assertFalseM(t, b.Contains(Coordinate{Lat: -0.5, Lng: -1.5}), "West")
	assertFalseM(t, b.Contains(Coordinate{Lat: -1.5, Lng: -2.5}), "South")
	assertFalseM(t, b.Contains(Coordinate{Lat: -1.5, Lng: -0.5}), "North")
}

func TestContainsSE(t *testing.T) {
	b := NewBoundingBox(
		Coordinate{Lat: 1.0, Lng: -2.0},
		Coordinate{Lat: 2.0, Lng: -1.0})
	
	assertTrueM(t, b.Contains(Coordinate{Lat: 1.5, Lng: -1.5}), "In Box")

	assertFalseM(t, b.Contains(Coordinate{Lat: 0.5, Lng: -1.5}), "East")
	assertFalseM(t, b.Contains(Coordinate{Lat: 2.5, Lng: -1.5}), "West")
	assertFalseM(t, b.Contains(Coordinate{Lat: 1.5, Lng: -2.5}), "South")
	assertFalseM(t, b.Contains(Coordinate{Lat: 1.5, Lng: -0.5}), "North")
}

func TestContainsAround0Latitude(t *testing.T) {
	b := NewBoundingBox(
		Coordinate{Lat: -1.0, Lng: 1.0},
		Coordinate{Lat: 1.0, Lng: 2.0})
	
	assertTrueM(t, b.Contains(Coordinate{Lat: 0, Lng: 1.5}), "In Box")

	assertFalseM(t, b.Contains(Coordinate{Lat: -1.5, Lng: 1.5}), "East")
	assertFalseM(t, b.Contains(Coordinate{Lat: 2.5, Lng: 1.5}), "West")
}

func TestContainsAround180Latitude(t *testing.T) {
	b := NewBoundingBox(
		Coordinate{Lat: 179.0, Lng: 1.0},
		Coordinate{Lat: -179.0,	Lng: 2.0})
	
	assertTrueM(t, b.Contains(Coordinate{Lat: -179.9, Lng: 1.5}), "In Box (E)")
	assertTrueM(t, b.Contains(Coordinate{Lat: 179.9, Lng: 1.5}), "In Box (W)")

	assertFalseM(t, b.Contains(Coordinate{Lat: 178, Lng: 1.5}), "East")
	assertFalseM(t, b.Contains(Coordinate{Lat: -178, Lng: 1.5}), "West")
}


func assertTrueM(t *testing.T, cond bool, msg string) {
	if !cond {
		t.Fatal(msg)
	}
}

func assertTrue(t *testing.T, cond bool) {
	assertTrueM(t, cond, "")
}

func assertFalseM(t *testing.T, cond bool, msg string) {
	if cond {
		t.Fatal(msg)
	}
}

func assertFalse(t *testing.T, cond bool) {
	assertTrueM(t, cond, "")
}
