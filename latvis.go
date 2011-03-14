package main

import (
	"flag"
	"fmt"
	"./latitude_api"
	"./latitude_xml"
	"./location"
	"log"
	"./tokens"
  "./visualizer"
)

func GetLocalHistorySource() *latitude_xml.FileSet {
	return latitude_xml.New("/home/mrjones/src/latvis/data")
}

func GetApiHistorySource() *latitude_api.AuthorizedConnection {
	connection := latitude_api.NewConnection()
	tokenStore := tokens.NewTokenStorage("tokens.txt")

  tokenSource := latitude_api.NewCachingTokenSource(connection, tokenStore);

	fmt.Println("User to generate map for:")
	var user string
	fmt.Scanln(&user)

	accessToken, err := tokenSource.GetToken(user);
	if err != nil{ log.Exit(err) }
	return connection.Authorize(accessToken);
}

func main() {
	var imageSize *int = flag.Int("imageSize", 720, "Size of resulting image")
	var useApi *bool = flag.Bool("useApi", true, "Use the API or local files")
	flag.Parse()

	var historySource location.HistorySource
	if *useApi {
		historySource = GetApiHistorySource()
	} else {
		historySource = GetLocalHistorySource()
	}
  vis := visualizer.NewVisualizer(*imageSize, &historySource);
  vis.GenerateImage("./vis.png");
}
