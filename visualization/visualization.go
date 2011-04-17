package visualization

import (
	"github.com/mrjones/latvis/location"

	"fmt"
	"image"
	"math"
)

type Heatmap struct {
	Points [][]float64
}

func scaleHeat(input int) float64 {
	return float64(math.Sqrt(math.Sqrt(float64(input))))
}

func HeatmapToImage(heatmap *Heatmap) image.Image {
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

func LocationHistoryAsHeatmap(history *location.History, size int, bounds *location.BoundingBox) *Heatmap {
	heatmap := &Heatmap{}
	heatmap.Points = make([][]float64, size, size)
	for i := 0 ; i < size ; i++ {
		heatmap.Points[i] = make([]float64, size, size)
	}

	if history.Len() < 1 {
		fmt.Println("Problem, Len() == 0")
	}

	initialized := false;
	maxX := 0.0
	minX := 0.0

	maxY := 0.0
	minY := 0.0

	for i := 0; i < history.Len(); i++ {
		if bounds.Contains(history.At(i)) {
			if !initialized || history.At(i).Lng < minX {
				minX = history.At(i).Lng
			}
			if !initialized || history.At(i).Lng > maxX {
				maxX = history.At(i).Lng
			}
			if !initialized || history.At(i).Lat < minY {
				minY = history.At(i).Lat
			}
			if !initialized || history.At(i).Lat > maxY {
				maxY = history.At(i).Lat
			}
			initialized = true
		}
	}

	fmt.Printf("xrange %f %f\n", minX, maxX)
	fmt.Printf("yrange %f %f\n", minY, maxY)

	counts := make([][]int, size, size)
	for i := 0 ; i < size ; i++ {
		counts[i] = make([]int, size, size)
	}

	xScale := float64(size-1) / (maxX - minX)
	yScale := float64(size-1) / (maxY - minY)
	scale := xScale
	if yScale < xScale {
		scale = yScale
	}

	for i := 0; i < history.Len(); i++ {
		if bounds.Contains(history.At(i)) {
			xBucket := int((history.At(i).Lng - minX) * scale)
			yBucket := size - int((history.At(i).Lat - minY)*scale) - 1
			counts[xBucket][yBucket]++
		}
	}

	maxCount := float64(0.0)
	for x := 0; x < len(counts); x++ {
		for y := 0; y < len(counts[x]); y++ {
			if scaleHeat(counts[x][y]) > maxCount {
				maxCount = scaleHeat(counts[x][y])
				fmt.Printf("counts[%d][%d] = %d\n", x, y, scaleHeat(counts[x][y]))
			}
		}
	}

	fmt.Printf("max: %d\n", maxCount)

	for x := 0; x < len(counts); x++ {
		for y := 0; y < len(counts[x]); y++ {
			heatmap.Points[x][y] = scaleHeat(counts[x][y]) / float64(maxCount)
		}
	}

	return heatmap
}
