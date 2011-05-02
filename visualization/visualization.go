package visualization

import (
	"github.com/mrjones/latvis/location"

	"fmt"
	"image"
	"log"
	"math"
	"os"
)

type Renderer interface {
	Render(grid *Grid, imageWidth int, imageHeight int) (image.Image, os.Error)
}

type Grid struct {
	points [][]int
	width int
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
	if x >= len(g.points) {
		fmt.Printf("x is too big: %d, official: %d, actual: %d\n", x, g.Width(), len(g.points))
	}
	if y >= len(g.points[x]) {
		fmt.Printf("y is too big: %d\n", y)
	}
	if (g.width != len(g.points)) {
		log.Fatalf("a internally mismatched len %d %d\n", g.width, len(g.points))
	}
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

func MakeImage(history *location.History, bounds *location.BoundingBox, width int, height int, renderer Renderer) (image.Image, os.Error) {
	grid := aggregateHistory(history, bounds, width, height)
	return renderer.Render(grid, width, height)
}

func aggregateHistory(history *location.History, bounds *location.BoundingBox, gridWidth int, gridHeight int) *Grid {
	grid := NewGrid(gridWidth, gridHeight)

	// For now, we always generate a square output image
	// but the selected box probably isn't exactly square.
	// As a result we won't want to fill the entirety of one
	// of the dimensions, or the picture will look stretched.
	// Figure out which dimension to constrict, and how much
	// to construct it by.

	inputSkew := bounds.Width() / bounds.Height();
	outputSkew := float64(gridWidth) / float64(gridHeight)
	xScale := 1.0
	yScale := 1.0
	// change 1.0 to gridWidth / gridHeight
	if (inputSkew >= outputSkew) {
		yScale = outputSkew / inputSkew
	} else {
		xScale = inputSkew / outputSkew
	}

	for i := 0; i < history.Len(); i++ {
		if bounds.Contains(history.At(i)) {
			xBucket := int(bounds.WidthFraction(history.At(i)) * xScale * float64(gridWidth))
			yBucket := int(bounds.HeightFraction(history.At(i)) * yScale * float64(gridHeight))
			yBucket = gridHeight - yBucket - 1
			grid.Inc(xBucket, yBucket)
		}
	}

	return grid
}

/////////////////////////////////////////////

func scaleHeat(input int) float64 {
	return float64(math.Sqrt(math.Sqrt(float64(input))))
}

type Heatmap struct {
	Points [][]float64
}

type BWRenderer struct {
}

func (r *BWRenderer) Render(grid *Grid, width int, height int) (image.Image, os.Error) {
	heatmap := gridAsHeatmap(grid, width, height)
	return heatmapToBWImage(heatmap), nil
}

func heatmapToBWImage(heatmap *Heatmap) image.Image {
	size := len(heatmap.Points)
	img := image.NewNRGBA(size, size)

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			val := heatmap.Points[i][j]
			if val > 0 {
				img.Pix[j*img.Stride+i] = image.NRGBAColor{uint8(0), uint8(0), uint8(0), 255}
			} else {
				img.Pix[j*img.Stride+i] = image.NRGBAColor{uint8(255), uint8(255), uint8(255), 255}
			}
		}
	}
	return img
}

func gridAsHeatmap(grid *Grid, width int, height int) *Heatmap {
	heatmap := &Heatmap{}
	heatmap.Points = make([][]float64, width)
	for i := 0 ; i < height ; i++ {
		heatmap.Points[i] = make([]float64, height)
	}

	maxCount := float64(0.0)
	for x := 0; x < grid.Width(); x++ {
		for y := 0; y < grid.Height(); y++ {
			if scaleHeat(grid.Get(x, y)) > maxCount {
				maxCount = scaleHeat(grid.Get(x, y))
			}
		}
	}

	for x := 0; x < grid.Width() ; x++ {
		for y := 0; y < grid.Height() ; y++ {
			heatmap.Points[x][y] = scaleHeat(grid.Get(x, y)) / float64(maxCount)
		}
	}

	return heatmap
}
