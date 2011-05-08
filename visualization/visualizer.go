package visualization

import (
	"github.com/mrjones/latvis/location"

	"bytes"
	"image"
	"image/png"
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
