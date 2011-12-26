package visualization

import (
	"github.com/mrjones/latvis/location"

	"bytes"
	"image"
	"image/png"
	"math"
	"os"
	"time"
)

type Visualizer struct {
  imageSize int // Generates a square image, each side is length "imageSize"
  historySource *location.HistorySource
	bounds *location.BoundingBox
	start time.Time
	end time.Time
}

func NewVisualizer(
		imageSize int,
	  historySource *location.HistorySource,
	  bounds *location.BoundingBox,
	  start time.Time,
	  end time.Time) *Visualizer {
  return &Visualizer{imageSize: imageSize, historySource: historySource, bounds: bounds, start: start, end: end};
}

func (v *Visualizer) Bytes() (*[]byte, os.Error) {
	history, err := readData(*v.historySource, v.start, v.end)
	if err != nil {
		return nil, err
	}

	renderer := &BWRenderer{}
	img, err := MakeImage(history, v.bounds, v.imageSize, v.imageSize, renderer)
	return renderImageToBytes(img)
}


func readData(historySource location.HistorySource, start time.Time, end time.Time) (*location.History, os.Error) {
	return historySource.FetchRange(start, end)
}

func renderImageToBytes(img image.Image) (*[]byte, os.Error) {
	buffer := bytes.NewBuffer(make([]byte, 0))

	if err := png.Encode(buffer, img); err != nil {
		return nil, err
	}

	bytes := buffer.Bytes()
	return &bytes, nil
}

// Interface for different styles of renderers to implement.
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
type Renderer interface {
	Render(grid *Grid, imageWidth int, imageHeight int) (image.Image, os.Error)
}

// Interface for callers to visualize a location history.
//
// history: 
// all the points in the users history
//
// bounds:
// a bounding box for selecting points to display.  points which fall outside
// the box will be discarded. The grid for rendering will also be constructed
// relative to this box (meaning white space outside of any points, but inside
// the bounding box will be preserved in the grid ... and probably by the
// Renderer as well).
//
// width / height:
// the width and height of the grid (how many cells to divide it into)
func MakeImage(history *location.History, bounds *location.BoundingBox, width int, height int, renderer Renderer) (image.Image, os.Error) {
	grid := aggregateHistory(history, bounds, width, height)
	return renderer.Render(grid, width, height)
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

//
// BWRenderer
//

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
//				img.Pix[j*img.Stride+i] = image.NRGBAColor{uint8(0), uint8(0), uint8(0), 255}
				img.Set(i, j, image.NRGBAColor{uint8(0), uint8(0), uint8(0), 255})
			} else {
//				img.Pix[j*img.Stride+i] = image.NRGBAColor{uint8(255), uint8(255), uint8(255), 255}
				img.Set(i, j, image.NRGBAColor{uint8(255), uint8(255), uint8(255), 255})
			}
		}
	}
	return img
}

//
// BWVectorRenderer
//

//type BWVectorRenderer struct {
//}
//
//func (r *BWVectorRenderer) Render(grid *Grid, width int, height int) (image.Image, os.Error) {
//	heatmap := gridAsHeatmap(grid, width, height)
//	return heatmapToBWVectorImage(heatmap), nil
//}


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
