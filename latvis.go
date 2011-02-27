package main

import (
	"fmt"
	"image"
	"image/png"
	"./latitude_api"
	"./latitude_xml"
	"./location"
	"log"
	oauth "github.com/hokapoka/goauth"
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

func GetAccessToken(user string, apiConnection *latitude_api.Connection, cache *tokens.Storage) (*oauth.AccessToken, os.Error) {

	accessToken, err := cache.Fetch(user)
	if err != nil{ return nil, err }
	if accessToken == nil {
		fmt.Printf("No saved token found. Generating new one")
		accessToken, err = apiConnection.NewAccessToken()
		if err != nil{ return nil, err }
		err = cache.Store(user, accessToken)
		if err != nil{ return nil, err }
	}
	return accessToken, nil
}

func GetLocalHistorySource() *latitude_xml.FileSet {
	return latitude_xml.New("/home/mrjones/src/latvis/data")
}

func GetApiHistorySource() *latitude_api.AuthorizedConnection {
	connection := latitude_api.NewConnection()
	tokenStore := tokens.NewTokenStorage("tokens.txt")

	fmt.Println("User to generate map for:")
	var user string
	fmt.Scanln(&user)

	accessToken, err := GetAccessToken(user, connection, tokenStore)
	if err != nil{ log.Exit(err) }
	return connection.Authorize(accessToken);
}

func main() {
	useApi := true
	var historySource location.HistorySource
	if useApi {
		historySource = GetApiHistorySource()
	} else {
		historySource = GetLocalHistorySource()
	}
	history := readData(historySource)
	size := 800
	img := visualization.HeatmapToImage(visualization.LocationHistoryAsHeatmap(history, size));
	renderImage(img, "vis.png")
}
