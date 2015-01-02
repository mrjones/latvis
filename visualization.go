package latvis

import (
	"bytes"
	"fmt"
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

// ======================================
// =========== SVG VISUALIZER ===========
// ======================================

type SvgVisualizer struct{}

func histogram(grid *Grid, resolution int64) {
	max := int64(0)
	sum := int64(0)
	count := int64(0)
	for x := 0 ; x < grid.Width(); x++ {
		for y := 0 ; y < grid.Height(); y++ {
			d := int64(grid.Get(x,y))
			if d > max {
				max = d
			}
			if d > 0 {
				count++
			}
			sum += d
		}
	}

	fmt.Printf("Max: %d\n", max)
	fmt.Printf("Sum: %d\n", sum)

	numBuckets := (max / resolution) + 1
	buckets := make([]int64, numBuckets)
	fmt.Printf("NumBuckets: %d\n", numBuckets)
	for x := 0 ; x < grid.Width(); x++ {
		for y := 0 ; y < grid.Height(); y++ {
			d := int64(grid.Get(x,y))
			if d > 0 {
//				fmt.Printf("%d -> %d\n", d, d/resolution)
				buckets[d / resolution]++
			}
		}
	}


	acc := int64(0)
	for i := int64(0) ; i < numBuckets; i++ {
		acc += buckets[i]
		if buckets[i] > 0 {
			fmt.Printf("%4d - %4d: %4d (%2.2f)\n", i * resolution, (i + 1) * resolution - 1, buckets[i], 100.0 * float64(acc) / float64(count))
		}
	}
}

func (s *SvgVisualizer) Visualize(history *History, bounds *BoundingBox, width int, height int) (*[]byte, error) {
	grid := aggregateHistory(history, bounds, width, height)
	histogram(grid, 5)

	var buf bytes.Buffer

	buf.WriteString("<svg xmlns=\"http://www.w3.org/2000/svg\" version=\"1.1\">")
	for x := 0 ; x < grid.Width(); x++ {
		for y := 0 ; y < grid.Height(); y++ {
			if grid.Get(x,y) > 0 {
//				r := (float64(grid.Get(x,y)) / 10) + 3
//				if (r < 3) { r = 3 }
//				if (r > 5) { r = 5 }

				cnt := grid.Get(x,y)
				r := 5 / (1 + math.Pow(math.E, -(float64(cnt)/10.0)))
				
				buf.WriteString(fmt.Sprintf(
					"<circle cx=\"%d\" cy=\"%d\" r=\"%.1f\" style=\"fill:rgb(0,0,0);\"/>", x * 10 , y * 10, r));
//				buf.WriteString(fmt.Sprintf(
//				"<rect x=\"%d\" y=\"%d\" width=\"3\" height=\"3\" style=\"fill:rgb(99,99,99);stroke-width:1;stroke:rgb(0,0,0)\" />", x * 3 , y * 3));
			}
		}
	}
	buf.WriteString("</svg>")

	dataBytes := buf.Bytes();
	return &dataBytes, nil
}
