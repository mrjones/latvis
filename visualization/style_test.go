package visualization

import (
	"github.com/mrjones/gt"

	"github.com/mrjones/latvis/location"

	"fmt"
	"image"
	"testing"
)

var (
	B = BLACK
	W = WHITE
)

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
	assertImage(t, [][]image.Color{
		[]image.Color{ W, W },
		[]image.Color{ B, W }}, img)

	h = append(h, &location.Coordinate{Lat: 1.5, Lng: 1.5})
	
	img, err = styler.Style(&h, bounds, 2, 2)
	gt.AssertNil(t, err)
	assertImage(t, [][]image.Color{
		[]image.Color{ W, B },
		[]image.Color{ B, W }}, img)
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
	assertImage(t, [][]image.Color{
		[]image.Color{ W, W, B, W, W },
		[]image.Color{ W, B, W, B, W },
		[]image.Color{ B, W, W, W, B }}, img)
}

func assertImage(t *testing.T, expected [][]image.Color, actual image.Image) {
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
	for x := 0 ; x < len(expected[0]) ; x++ {
		for y := 0 ; y < len(expected) ; y++ {
			expPix := expected[y][x]
			gt.AssertEqualM(t, expPix, actual.At(x, y),
				fmt.Sprintf("Unexpected pixel at: (%d, %d)", x, y))
		}
	}
}
