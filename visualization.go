package latvis

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
)

// ======================================
// ========== VISUALIZATION API =========
// ======================================

// Turns a location history into a representation of that history
// - history:   The list of points to render.
// - bounds:    The borders of the image in latitude/longitude
// - width/height:  The width & height of the final image in pixels
//
// returns
// - a []byte representing a PNG image
//
// TODO(mrjones): return a ContentType along with []bytes
// TODO(mrjones): do width & height make sense for non-PNG return types?
type Visualizer interface {
	Visualize(history *History,
		bounds *BoundingBox,
		imageWidth,
		imageHeight int) (*[]byte, error)
}

// ======================================
// ========== BW PNG VISUALIZER =========
// ======================================

var (
	BLACK = color.NRGBA{uint8(0), uint8(0), uint8(0), 255}
	WHITE = color.NRGBA{uint8(255), uint8(255), uint8(255), 255}
)

type IntensityGrid struct {
	Points [][]float64
}

type BwPngVisualizer struct{}

func (r *BwPngVisualizer) Visualize(history *History, bounds *BoundingBox, width int, height int) (*[]byte, error) {
	return imageToPNGBytes(r.makeImage(history, bounds, width, height))
}

// Seam for testing
func (r *BwPngVisualizer) makeImage(history *History, bounds *BoundingBox, width int, height int) image.Image {
	grid := aggregateHistory(history, bounds, width, height)
	intensityGrid := formatAsIntensityGrid(grid, width, height)
	return intensityGridToBWImage(intensityGrid)
}

func aggregateHistory(history *History, bounds *BoundingBox, gridWidth int, gridHeight int) *Grid {
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

func scaleHeat(input int) float64 {
	return float64(math.Sqrt(math.Sqrt(float64(input))))
}

func formatAsIntensityGrid(grid *Grid, width int, height int) *IntensityGrid {
	intensityGrid := &IntensityGrid{}
	intensityGrid.Points = make([][]float64, width)
	for i := 0; i < width; i++ {
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

func intensityGridToBWImage(intensityGrid *IntensityGrid) image.Image {
	width := len(intensityGrid.Points)
	height := len(intensityGrid.Points[0])
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
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

func imageToPNGBytes(img image.Image) (*[]byte, error) {
	buffer := bytes.NewBuffer(make([]byte, 0))

	if err := png.Encode(buffer, img); err != nil {
		return nil, err
	}

	bytes := buffer.Bytes()
	return &bytes, nil
}

// ======================================
// ========= GRID (HELPER CLASS) ========
// ======================================

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
