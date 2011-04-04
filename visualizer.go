package visualizer

import (
	"image"
	"image/png"
	"./location"
	"log"
	"os"
	"./visualization"
)

type Visualizer struct {
  imageSize int // Generates a square image, each side is length "imageSize"
  historySource *location.HistorySource;
}

func NewVisualizer(imageSize int, historySource *location.HistorySource) *Visualizer {
  return &Visualizer{imageSize: imageSize, historySource: historySource};
}

func (v *Visualizer) GenerateImage(path string) {
	history := readData(*v.historySource)
	img := visualization.HeatmapToImage(
      visualization.LocationHistoryAsHeatmap(history, v.imageSize));
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
