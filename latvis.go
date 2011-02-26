package main

import (
	"image"
	"image/png"
	"./latitude_xml"
	"./location"
	"log"
	"os"
	"./visualization"
)

func readAndAppendData(filename string, history *location.History) {
	xmlFile := latitude_xml.New(filename)
	localHistory, err := xmlFile.GetHistory()
	if err != nil { log.Exit(err) }
	history.AddAll(localHistory)
}

func readData(filenames []string) *location.History {
	history := &location.History{}
	for i := 0 ; i < len(filenames) ; i++ {
		readAndAppendData(filenames[i], history)
	}

	return history
}

func renderImage(img image.Image, filename string) {
	f, err := os.Open(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Exit(err)
	}
	if err = png.Encode(f, img); err != nil {
		log.Exit(err)
	}
}

func main() {
	size := 300

	datafiles := [...]string{
		"/home/mrjones/src/latvis/data/2010-07.kml",
		"/home/mrjones/src/latvis/data/2010-08.kml",
		"/home/mrjones/src/latvis/data/2010-09.kml",
		"/home/mrjones/src/latvis/data/2010-10.kml",
		"/home/mrjones/src/latvis/data/2010-11.kml",
		"/home/mrjones/src/latvis/data/2010-12.kml",
		"/home/mrjones/src/latvis/data/jan2011.kml",
		"/home/mrjones/src/latvis/data/feb2011.kml",
	}

	history := readData(datafiles[:])
	img := visualization.HeatmapToImage(visualization.LocationHistoryAsHeatmap(history, size));
	renderImage(img, "vis.png")
}
