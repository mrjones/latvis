package latvis

import (
	"github.com/mrjones/gt"

	"fmt"
	"image"
	"image/color"
	"testing"
)

var (
	B = BLACK
	W = WHITE
)

func TestBWStyler2By2(t *testing.T) {
	bounds, err := NewBoundingBox(
		Coordinate{Lat: 0, Lng: 0},
		Coordinate{Lat: 2, Lng: 2})
	gt.AssertNil(t, err)

	h := make(History, 0)
	h = append(h, &Coordinate{Lat: .5, Lng: .5})

	styler := &BWStyler{} 
	img := styler.makeImage(&h, bounds, 2, 2)
	assertImage(t, [][]color.Color{
		[]color.Color{W, W},
		[]color.Color{B, W}}, img)

	h = append(h, &Coordinate{Lat: 1.5, Lng: 1.5})

	img = styler.makeImage(&h, bounds, 2, 2)
	assertImage(t, [][]color.Color{
		[]color.Color{W, B},
		[]color.Color{B, W}}, img)
}

func TestBWStylerNotSquare(t *testing.T) {
	bounds, err := NewBoundingBox(
		Coordinate{Lat: 0, Lng: 0},
		Coordinate{Lat: 3, Lng: 5})
	gt.AssertNil(t, err)

	h := make(History, 0)
	h = append(h, &Coordinate{Lat: .5, Lng: .5})
	h = append(h, &Coordinate{Lat: 1.5, Lng: 1.5})
	h = append(h, &Coordinate{Lat: 2.5, Lng: 2.5})
	h = append(h, &Coordinate{Lat: 1.5, Lng: 3.5})
	h = append(h, &Coordinate{Lat: .5, Lng: 4.5})

	styler := &BWStyler{}

	img := styler.makeImage(&h, bounds, 5, 3)
	assertImage(t, [][]color.Color{
		[]color.Color{W, W, B, W, W},
		[]color.Color{W, B, W, B, W},
		[]color.Color{B, W, W, W, B}}, img)
}

func TestBWStylerSmushed(t *testing.T) {
	bounds, err := NewBoundingBox(
		Coordinate{Lat: 0, Lng: 0},
		Coordinate{Lat: 50, Lng: 50})
	gt.AssertNil(t, err)

	h := make(History, 0)
	// Lots of points, but they're all in the lower left.
	h = append(h, &Coordinate{Lat: 1, Lng: 1})
	h = append(h, &Coordinate{Lat: 2, Lng: 2})
	h = append(h, &Coordinate{Lat: 3, Lng: 3})
	h = append(h, &Coordinate{Lat: 4, Lng: 4})
	h = append(h, &Coordinate{Lat: 5, Lng: 5})

	styler := &BWStyler{}

	img := styler.makeImage(&h, bounds, 2, 2)
	assertImage(t, [][]color.Color{
		[]color.Color{W, W},
		[]color.Color{B, W}}, img)

}

func assertImage(t *testing.T, expected [][]color.Color, actual image.Image) {
	gt.AssertEqualM(t, len(expected), actual.Bounds().Dy(), "Unexpected image height")
	gt.AssertEqualM(t, len(expected[0]), actual.Bounds().Dx(), "Unexpected image width")

	// WxH image (0,0) is in the top left:
	// (i.e. increasing Y-coordinates move downwards)
	// (0,0) ----- (W,0)
	//   |           |
	// (0,H) ----- (W,H)
	//
	// This works, because the ASCII arrays in the asserts are laid out similarly.
	// However, it means that printed coordinates might be y-inverted from
	// intuition.
	for x := 0; x < len(expected[0]); x++ {
		for y := 0; y < len(expected); y++ {
			expPix := expected[y][x]
			gt.AssertEqualM(t, expPix, actual.At(x, y),
				fmt.Sprintf("Unexpected pixel at: (%d, %d)", x, y))
		}
	}
}
