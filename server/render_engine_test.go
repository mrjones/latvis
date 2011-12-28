package server

import (
	"github.com/mrjones/gt"
	"github.com/mrjones/latvis/location"

	"testing"
)

func TestSquareBox(t *testing.T) {
	box, err := location.NewBoundingBox(
		location.Coordinate{Lat: 0, Lng: 0},
		location.Coordinate{Lat: 10, Lng: 10})
	gt.AssertNil(t, err)

	w, h := imgSize(box, 500)

	gt.AssertEqualM(t, 500, w, "Width should be maxed for a square box")
	gt.AssertEqualM(t, 500, h, "Height should be maxed for a square box")
}
