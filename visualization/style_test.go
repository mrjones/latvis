package visualization

import (
	"github.com/mrjones/gt"

	"github.com/mrjones/latvis/location"

	"fmt"
	"image"
	"testing"
)

//
// WxH image (0,0) is in the top left:
// (i.e. increasing Y-coordinates move downwards)
// (0,0) ----- (W,0)
//   |           |
// (0,H) ----- (W,H)
//
// But in Lat/Lng, increasing Latitudes move upwards.
func TestBWStyler2By2(t *testing.T) {
	bounds, err := location.NewBoundingBox(
		location.Coordinate{Lat: 0, Lng: 0},
		location.Coordinate{Lat: 2, Lng: 2})

	gt.AssertNil(t, err)

	h := make(location.History, 0)

	h = append(h, &location.Coordinate{Lat: .5, Lng: .5})

	styler := &BWStyler{}

	img, err := styler.Style(&h, bounds, 2, 2)
	gt.AssertNil(t, err)
	assertImage(t, [][]string{
		[]string{ "W", "W" },
		[]string{ "B", "W" }}, img)

	h = append(h, &location.Coordinate{Lat: 1.5, Lng: 1.5})
	
	img, err = styler.Style(&h, bounds, 2, 2)
	gt.AssertNil(t, err)
	assertImage(t, [][]string{
		[]string{ "W", "B" },
		[]string{ "B", "W" }}, img)
}

func TestBWStylerNotSquare(t *testing.T) {
	bounds, err := location.NewBoundingBox(
		location.Coordinate{Lat: 0, Lng: 0},
		location.Coordinate{Lat: 3, Lng: 5})

	gt.AssertNil(t, err)

	h := make(location.History, 0)
	h = append(h, &location.Coordinate{Lat: .5, Lng: .5})
	h = append(h, &location.Coordinate{Lat: 1.5, Lng: 1.5})
	h = append(h, &location.Coordinate{Lat: 2.5, Lng: 2.5})
	h = append(h, &location.Coordinate{Lat: 1.5, Lng: 3.5})
	h = append(h, &location.Coordinate{Lat: .5, Lng: 4.5})

	styler := &BWStyler{}

	img, err := styler.Style(&h, bounds, 5, 3)
	gt.AssertNil(t, err)
	assertImage(t, [][]string{
		[]string{ "W", "W", "B", "W", "W" },
		[]string{ "W", "B", "W", "B", "W" },
		[]string{ "B", "W", "W", "W", "B" }}, img)
}

func assertImage(t *testing.T, expected [][]string, actual image.Image) {
	gt.AssertEqualM(t, len(expected), actual.Bounds().Dy(), "Unexpected image height")
	gt.AssertEqualM(t, len(expected[0]), actual.Bounds().Dx(), "Unexpected image width")

	for x := 0 ; x < len(expected[0]) ; x++ {
		for y := 0 ; y < len(expected) ; y++ {
			expPix := expected[y][x]
			if expPix == "B" { assertBlack(t, actual, x, y) }
			if expPix == "W" { assertWhite(t, actual, x, y) }
		}
	}
}

func assertBlack(t *testing.T, i image.Image, x, y int) {
	r, g, b, a := i.At(x, y).RGBA()
	gt.AssertEqualM(t, uint32(0), r,
		fmt.Sprintf("Red should be 0 for black at (%d, %d).", x, y))
	gt.AssertEqualM(t, uint32(0), g, "Blue should be 0 for black")
	gt.AssertEqualM(t, uint32(0), b, "Green should be 0 for black")
	gt.AssertEqualM(t, uint32(0xFFFF), a, "Alpha should be max-uint32")
}

func assertWhite(t *testing.T, i image.Image, x, y int) {
	r, g, b, a := i.At(x, y).RGBA()
	gt.AssertEqualM(t, uint32(0xFFFF), r,
		fmt.Sprintf("Red should be max-uint32 for white at (%d, %d).", x, y))
	gt.AssertEqualM(t, uint32(0xFFFF), g, "Blue should be max-uint32 for white")
	gt.AssertEqualM(t, uint32(0xFFFF), b, "Green should be max-uint32 for white")
	gt.AssertEqualM(t, uint32(0xFFFF), a, "Alpha should be max-uint32")
}
