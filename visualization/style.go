package visualization

import (
	"image"
	"math"
	"os"
)

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
