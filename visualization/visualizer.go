package visualization

import (
	"github.com/mrjones/latvis/location"

	"image"
	"image/png"
	"log"
	"os"
)

type Visualizer struct {
  imageSize int // Generates a square image, each side is length "imageSize"
  historySource *location.HistorySource;
}

func NewVisualizer(imageSize int, historySource *location.HistorySource) *Visualizer {
  return &Visualizer{imageSize: imageSize, historySource: historySource};
}

func (v *Visualizer) GenerateImage(path string) {
	bounds, err := location.NewBoundingBox(
		location.Coordinate{Lat: -74.02, Lng: 40.703},
		location.Coordinate{Lat: -73.96, Lng: 40.8})

	if err != nil {
		// Propogate?
		log.Fatal(err)
	}

	history := readData(*v.historySource)
	img := HeatmapToImage(
      LocationHistoryAsHeatmap(history, v.imageSize, bounds));
	renderImage(img, path)
}

///////

func readAndAppendData(source location.HistorySource, year int64, month int, history *location.History) {
	localHistory, err := source.GetHistory(year, month)
	if err != nil { log.Fatal(err) }
	history.AddAll(localHistory)
}

func readData(historySource location.HistorySource) *location.History {
	history := &location.History{}
	readAndAppendData(historySource, 2010, 7, history)
	readAndAppendData(historySource, 2010, 8, history)
	readAndAppendData(historySource, 2010, 9, history)
	readAndAppendData(historySource, 2010, 10, history)
	readAndAppendData(historySource, 2010, 11, history)
	readAndAppendData(historySource, 2010, 12, history)
	readAndAppendData(historySource, 2011, 1, history)
	readAndAppendData(historySource, 2011, 2, history)
	readAndAppendData(historySource, 2011, 3, history)
	readAndAppendData(historySource, 2011, 4, history)

	return history
}

func renderImage(img image.Image, filename string) {
	f, err := os.Open(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	if err = png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}
