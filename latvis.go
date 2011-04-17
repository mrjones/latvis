package main

import (
	"github.com/mrjones/latvis/latitude"
	"github.com/mrjones/latvis/location"
	"github.com/mrjones/latvis/visualization"

	"flag"
	"fmt"
	"log"
)

//func GetLocalHistorySource() *latitude_xml.FileSet {
//	return latitude_xml.New("/home/mrjones/src/latvis/data")
//}

func GetApiHistorySource() *latitude.AuthorizedConnection {
	connection := latitude.NewConnection()
	tokenStore := latitude.NewTokenStorage("tokens.txt")

  tokenSource := latitude.NewCachingTokenSource(connection, tokenStore);

	fmt.Println("User to generate map for:")
	var user string
	fmt.Scanln(&user)

	accessToken, err := tokenSource.GetToken(user);
	if err != nil{ log.Fatal(err) }
	return connection.Authorize(accessToken);
}

func main() {
	var imageSize *int = flag.Int("imageSize", 720, "Size of resulting image")
	flag.Parse()

	var historySource location.HistorySource
//	if *useApi {
		historySource = GetApiHistorySource()
//	} else {
//		historySource = GetLocalHistorySource()
//	}
  vis := visualization.NewVisualizer(*imageSize, &historySource);
  vis.GenerateImage("./vis.png");
}
