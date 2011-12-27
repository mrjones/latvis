package visualization

import (
	"github.com/mrjones/latvis/location"

	"bytes"
	"image"
	"image/png"
	"os"
)

// Interface to implement different styles of maps.
type Styler interface {
	Style(history *location.History, bounds *location.BoundingBox, imageWidth int, imageHeight int) (image.Image, os.Error)
}


// Turns a location history into an image, based on the selected style.
// - history:   The list of points to render.
// - bounds:    The borders of the image (points outside the bounds are dropped).
// - styler:    The Styler to use when turning the history into an image
// - imageSize: (Asumes a square) the width & height of the final image in pixels
//
// returns
// - a []byte representing a PNG image
func Draw(history *location.History, bounds *location.BoundingBox, styler Styler, imageSize int) (*[]byte, os.Error) {
	width := imageSize
	height := imageSize

	img, err := styler.Style(history, bounds, width, height)

	if err != nil {
		return nil, err
	}
	return imageToPNGBytes(img)
}

func imageToPNGBytes(img image.Image) (*[]byte, os.Error) {
	buffer := bytes.NewBuffer(make([]byte, 0))

	if err := png.Encode(buffer, img); err != nil {
		return nil, err
	}

	bytes := buffer.Bytes()
	return &bytes, nil
}

type Grid struct {
	points [][]int
	width  int
	height int
}

func NewGrid(width, height int) *Grid {
	grid := Grid{points: make([][]int, width, width)}
	for i, _ := range grid.points {
		grid.points[i] = make([]int, height, height)
	}
	grid.width = width
	grid.height = height
	return &grid
}

func (g *Grid) Get(x, y int) int {
	return g.points[x][y]
}

func (g *Grid) Set(x, y, val int) {
	g.points[x][y] = val
}

func (g *Grid) Inc(x, y int) {
	g.points[x][y]++
}

func (g *Grid) Width() int {
	return g.width
}

func (g *Grid) Height() int {
	return g.height
}

func aggregateHistory(history *location.History, bounds *location.BoundingBox, gridWidth int, gridHeight int) *Grid {
	grid := NewGrid(gridWidth, gridHeight)

	// For now, we always generate a square output image
	// but the selected box probably isn't exactly square.
	// As a result we won't want to fill the entirety of one
	// of the dimensions, or the picture will look stretched.
	// Figure out which dimension to constrict, and how much
	// to construct it by.

	inputSkew := bounds.Width() / bounds.Height()
	outputSkew := float64(gridWidth) / float64(gridHeight)
	xScale := 1.0
	yScale := 1.0

	if inputSkew >= outputSkew {
		yScale = outputSkew / inputSkew
	} else {
		xScale = inputSkew / outputSkew
	}

	for i := 0; i < history.Len(); i++ {
		if bounds.Contains(history.At(i)) {
			xBucket := int(bounds.WidthFraction(history.At(i)) * xScale * float64(gridWidth))
			yBucket := int(bounds.HeightFraction(history.At(i)) * yScale * float64(gridHeight))
			// TODO(mrjones): explain this
			yBucket = gridHeight - yBucket - 1
			grid.Inc(xBucket, yBucket)
		}
	}

	return grid
}
