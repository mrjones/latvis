package main

import (
	"fmt"
	"image"
	"image/png"
	"./latitude_api"
	"./latitude_xml"
	"./location"
	"log"
	"os"
	"./tokens"
	"./visualization"
)

func readAndAppendData(source location.HistorySource, year int64, month int, history *location.History) {
	localHistory, err := source.GetHistory(year, month)
	if err != nil { log.Exit(err) }
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

	xmlHistorySource := latitude_xml.New("/home/mrjones/src/latvis/data")


	connection := latitude_api.NewConnection()

	tokenStore := tokens.NewTokenStorage("tokens.txt")
	fmt.Println("User to generate map for:")
	var user string
	fmt.Scanln(&user)

	accessToken, err := tokenStore.Fetch(user)
	if err != nil{ log.Exit(err) }
	if accessToken == nil {
		fmt.Printf("No saved token found. Generating new one")
		accessToken, err = connection.NewAccessToken()
		if err != nil{ log.Exit(err) }
		err = tokenStore.Store(user, accessToken)
		if err != nil{ log.Exit(err) }
	}

	apiHistorySource := connection.Authorize(accessToken);

	useApi := true
	var history *location.History
	if useApi {
		history = readData(apiHistorySource)
	} else {
		history = readData(xmlHistorySource)
	}
	img := visualization.HeatmapToImage(visualization.LocationHistoryAsHeatmap(history, size));
	renderImage(img, "vis.png")
}
