package visualization

import (
	"github.com/mrjones/gt"

	"image"
	"testing"
)

func TestBWStyler2By2(t *testing.T) {
	g := NewGrid(2, 2)
	g.Set(0, 0, 1)

	styler := &BWStyler{}

	img, err := styler.Style(g, 2, 2)
	gt.AssertNil(t, err)
	assertBlack(t, img.At(0, 0))
	assertWhite(t, img.At(1, 0))
	assertWhite(t, img.At(0, 1))
	assertWhite(t, img.At(1, 1))

	g.Set(1, 1, 1)
	img, err = styler.Style(g, 2, 2)
	gt.AssertNil(t, err)
	assertBlack(t, img.At(0, 0))
	assertWhite(t, img.At(1, 0))
	assertWhite(t, img.At(0, 1))
	assertBlack(t, img.At(1, 1))
}

func TestBWStylerNotSquare(t *testing.T) {
	g := NewGrid(3, 10)
	g.Set(0, 0, 1)
	g.Set(1, 1, 1)
	g.Set(2, 2, 1)

	styler := &BWStyler{}

	img, err := styler.Style(g, 3, 10)
	gt.AssertNil(t, err)
	assertBlack(t, img.At(0, 0))
	assertBlack(t, img.At(1, 1))
	assertBlack(t, img.At(2, 2))

	assertWhite(t, img.At(1, 0))
	assertWhite(t, img.At(0, 1))
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
