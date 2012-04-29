package latvis

import (
	"image"
	"image/color"
	"math"
)

var (
	BLACK = color.NRGBA{uint8(0), uint8(0), uint8(0), 255}
	WHITE = color.NRGBA{uint8(255), uint8(255), uint8(255), 255}
)

//
// BWStyler
//

type IntensityGrid struct {
	Points [][]float64
}

type BWStyler struct{}

func (r *BWStyler) Style(history *History, bounds *BoundingBox, width int, height int) (image.Image, error) {
	grid := aggregateHistory(history, bounds, width, height)
	intensityGrid := formatAsIntensityGrid(grid, width, height)
	return intensityGridToBWImage(intensityGrid), nil
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

func scaleHeat(input int) float64 {
	return float64(math.Sqrt(math.Sqrt(float64(input))))
}