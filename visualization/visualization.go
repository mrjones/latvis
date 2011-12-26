package visualization

import (
	"github.com/mrjones/latvis/location"

	"bytes"
	"image"
	"image/png"
	"math"
	"os"
)

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

	grid := aggregateHistory(history, bounds, width, height)
	img, err := styler.Style(grid, width, height)

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

// Interface to implement different styles of maps.
//
// *Subject to change*
//
// The input is a "Grid" (see below), representing location history data:
// The entire selected are is broken down into a coordinate grid, with a
// discrete number of cells.  The Grid object represents the number of location
// history events occuring in each cell.
//
// TODO(mrjones): The grid is latitude / longitude independent. This is fine for
// rendering context-independent PNGs, but won't work if we want to show the data
// overlaid on a map.
//
// TODO(mrjones): what about returning a (byte[], mime-type)?
// that would let us handle images as well as other things like KML files for maps
type Styler interface {
	Style(grid *Grid, imageWidth int, imageHeight int) (image.Image, os.Error)
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

/////////////////////////////////////////////

//
// BWStyler
//

type IntensityGrid struct {
	Points [][]float64
}

type BWStyler struct { }

func (r *BWStyler) Style(grid *Grid, width int, height int) (image.Image, os.Error) {
	intensityGrid := formatAsIntensityGrid(grid, width, height)
	return intensityGridToBWImage(intensityGrid), nil
}

func intensityGridToBWImage(intensityGrid *IntensityGrid) image.Image {
	size := len(intensityGrid.Points)
	img := image.NewNRGBA(size, size)

	BLACK := image.NRGBAColor{uint8(0), uint8(0), uint8(0), 255}
	WHITE := image.NRGBAColor{uint8(255), uint8(255), uint8(255), 255}

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			val := intensityGrid.Points[i][j]
			if val > 0 {
				img.Set(i, j, BLACK)
			} else {
				img.Set(i, j, WHITE)
			}
		}
	}
	return img
}

func formatAsIntensityGrid(grid *Grid, width int, height int) *IntensityGrid {
	intensityGrid := &IntensityGrid{}
	intensityGrid.Points = make([][]float64, width)
	for i := 0; i < height; i++ {
		intensityGrid.Points[i] = make([]float64, height)
	}

	maxCount := float64(0.0)
	for x := 0; x < grid.Width(); x++ {
		for y := 0; y < grid.Height(); y++ {
			if scaleHeat(grid.Get(x, y)) > maxCount {
				maxCount = scaleHeat(grid.Get(x, y))
			}
		}
	}

	for x := 0; x < grid.Width(); x++ {
		for y := 0; y < grid.Height(); y++ {
			intensityGrid.Points[x][y] = scaleHeat(grid.Get(x, y)) / float64(maxCount)
		}
	}

	return intensityGrid
}

func scaleHeat(input int) float64 {
	return float64(math.Sqrt(math.Sqrt(float64(input))))
}
