package location

import (
	"github.com/mrjones/gt"

	"testing"
)

func TestContainsBoundaries(t *testing.T) {
	b, err := NewBoundingBox(
		Coordinate{Lat: -1.0, Lng: -1.0},
		Coordinate{Lat: 1.0, Lng: 1.0})
	
	gt.AssertNil(t, err)

	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: -1, Lng: 0}), "South")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 1, Lng: 0}), "North")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 0, Lng: -1}), "East")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 0, Lng: 1}), "West")
}

func TestContainsNE(t *testing.T) {
	b, err := NewBoundingBox(
		Coordinate{Lat: 1.0, Lng: 1.0},
		Coordinate{Lat: 2.0, Lng: 2.0})
	
	gt.AssertNil(t, err)

	gt.AssertTrueM(t, b.Contains(&Coordinate{Lat: 1.5, Lng: 1.5}), "In Box")

	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 0.5, Lng: 1.5}), "East")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 2.5, Lng: 1.5}), "West")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 1.5, Lng: 0.5}), "South")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 1.5, Lng: 2.5}), "North")
}

func TestContainsNW(t *testing.T) {
	b, err := NewBoundingBox(
		Coordinate{Lat: -2.0, Lng: 1.0},
		Coordinate{Lat: -1.0,	Lng: 2.0})
	
	gt.AssertNil(t, err)

	gt.AssertTrueM(t, b.Contains(&Coordinate{Lat: -1.5, Lng: 1.5}), "In Box")

	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: -2.5, Lng: 1.5}), "East")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: -0.5, Lng: 1.5}), "West")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: -1.5, Lng: 0.5}), "South")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: -1.5, Lng: 2.5}), "North")
}

func TestContainsSW(t *testing.T) {
	b, err := NewBoundingBox(
		Coordinate{Lat: -2.0, Lng: -2.0},
		Coordinate{Lat: -1, Lng: -1.0})

	gt.AssertNil(t, err)
	
	gt.AssertTrueM(t, b.Contains(&Coordinate{Lat: -1.5, Lng: -1.5}), "In Box")

	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: -2.5, Lng: -1.5}), "East")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: -0.5, Lng: -1.5}), "West")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: -1.5, Lng: -2.5}), "South")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: -1.5, Lng: -0.5}), "North")
}

func TestContainsSE(t *testing.T) {
	b, err := NewBoundingBox(
		Coordinate{Lat: 1.0, Lng: -2.0},
		Coordinate{Lat: 2.0, Lng: -1.0})
	
	gt.AssertNil(t, err)

	gt.AssertTrueM(t, b.Contains(&Coordinate{Lat: 1.5, Lng: -1.5}), "In Box")

	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 0.5, Lng: -1.5}), "East")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 2.5, Lng: -1.5}), "West")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 1.5, Lng: -2.5}), "South")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 1.5, Lng: -0.5}), "North")
}

func TestContainsAround0Latitude(t *testing.T) {
	b, err := NewBoundingBox(
		Coordinate{Lat: -1.0, Lng: 1.0},
		Coordinate{Lat: 1.0, Lng: 2.0})
	
	gt.AssertNil(t, err)

	gt.AssertTrueM(t, b.Contains(&Coordinate{Lat: 0, Lng: 1.5}), "In Box")

	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: -1.5, Lng: 1.5}), "East")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 2.5, Lng: 1.5}), "West")
}

func TestContainsAround180Longitude(t *testing.T) {
	b, err := NewBoundingBox(
		Coordinate{Lat: 1.0, Lng: 179.0},
		Coordinate{Lat: 2.0, Lng: -179.0})
	
	gt.AssertNil(t, err)

	gt.AssertTrueM(t, b.Contains(&Coordinate{Lat: 1.5, Lng: -179.9}), "In Box (E)")
	gt.AssertTrueM(t, b.Contains(&Coordinate{Lat: 1.5, Lng: 179.9}), "In Box (W)")

	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 1.5, Lng: 178}), "East")
	gt.AssertFalseM(t, b.Contains(&Coordinate{Lat: 1.5, Lng: -178}), "West")
}

func TestInvalidBox(t *testing.T) {
	_, err := NewBoundingBox(
		Coordinate{Lat: 2.0, Lng: 1.0},
		Coordinate{Lat: 1.0, Lng: 2.0})
	
	gt.AssertNotNil(t, err)
}

func TestNormalWidth(t *testing.T) {
	b, err := NewBoundingBox(
		Coordinate{Lat: 1.0, Lng: 1.0},
		Coordinate{Lat: 10.0, Lng: 2.0})

	gt.AssertNil(t, err)
	if b.Width() != 1.0 {
		t.Fatal("Wrong width: Expected 1.0, Actual: %f", b.Width())
	}
	wf := b.WidthFraction(&Coordinate{Lat: 1.0, Lng: 1.5}) 
	if wf != 0.5 {
		t.Fatal("Wrong width fraction: Expected .5, Actual: %f", wf)
	}
	wf = b.WidthFraction(&Coordinate{Lat: 1.0, Lng: 1.25}) 
	if wf != 0.25 {
		t.Fatal("Wrong width fraction: Expected .25, Actual: %f", wf)
	}

	if b.Height() != 9.0 {
		t.Fatal("Wrong height: Expected 9.0, Actual: %f", b.Height())
	}
	
}

func TestWidthAround180(t *testing.T) {
	b, err := NewBoundingBox(
		Coordinate{Lat: 1.0, Lng: 179.0},
		Coordinate{Lat: 10.0, Lng: -179.0})
	
	gt.AssertNil(t, err)
	if b.Width() != 2.0 {
		t.Fatalf("Wrong width: Expected 2.0, Actual: %f", b.Width())
	}
	wf := b.WidthFraction(&Coordinate{Lat: 1.0, Lng: 179.5}) 
	if wf != 0.25 {
		t.Fatalf("Wrong width fraction: Expected .25, Actual: %f", wf)
	}
	wf = b.WidthFraction(&Coordinate{Lat: 1.0, Lng: -179.5}) 
	if wf != 0.75 {
		t.Fatalf("Wrong width fraction: Expected .75, Actual: %f", wf)
	}
}
