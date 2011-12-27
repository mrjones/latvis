package visualization

import (
	"github.com/mrjones/gt"

	"github.com/mrjones/latvis/location"

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

	// Expected Image:
	//  W W
	//  B W
	img, err := styler.Style(&h, bounds, 2, 2)
	gt.AssertNil(t, err)
	assertWhite(t, img.At(0, 0))
	assertWhite(t, img.At(1, 0))
	assertBlack(t, img.At(0, 1))
	assertWhite(t, img.At(1, 1))

	h = append(h, &location.Coordinate{Lat: 1.5, Lng: 1.5})
	
	// Expected Image:
	//  W B
	//  B W
	img, err = styler.Style(&h, bounds, 2, 2)
	gt.AssertNil(t, err)
	assertWhite(t, img.At(0, 0))
	assertBlack(t, img.At(1, 0))
	assertBlack(t, img.At(0, 1))
	assertWhite(t, img.At(1, 1))
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

	styler := &BWStyler{}

	// Expected Image:
	//  W W B W W
	//  W B W W W
	//  B W W W W
	img, err := styler.Style(&h, bounds, 5, 3)
	gt.AssertNil(t, err)
	assertWhite(t, img.At(0, 0))

	assertBlack(t, img.At(0, 2))
	assertBlack(t, img.At(1, 1))
	assertBlack(t, img.At(2, 0))
}

func assertBlack(t *testing.T, c image.Color) {
	r, g, b, a := c.RGBA()
	gt.AssertEqualM(t, uint32(0), r, "Red should be 0 for black")
	gt.AssertEqualM(t, uint32(0), g, "Blue should be 0 for black")
	gt.AssertEqualM(t, uint32(0), b, "Green should be 0 for black")
	gt.AssertEqualM(t, uint32(0xFFFF), a, "Alpha should be max-uint32")
}

func assertWhite(t *testing.T, c image.Color) {
	r, g, b, a := c.RGBA()
	gt.AssertEqualM(t, uint32(0xFFFF), r, "Red should be max-uint32 for white")
	gt.AssertEqualM(t, uint32(0xFFFF), g, "Blue should be max-uint32 for white")
	gt.AssertEqualM(t, uint32(0xFFFF), b, "Green should be max-uint32 for white")
	gt.AssertEqualM(t, uint32(0xFFFF), a, "Alpha should be max-uint32")
}
