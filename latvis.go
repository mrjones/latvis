package main

import (
	"github.com/mrjones/latvis/latitude"
  "github.com/mrjones/latvis/location"
	"github.com/mrjones/latvis/server"
	"github.com/mrjones/latvis/visualization"

	"flag"
	"fmt"
	"log"
)

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
	var runAsServer *bool = flag.Bool("server", true, "Run as server (vs. one off)")
	flag.Parse()

  if *runAsServer {
    server.Serve()     
  } else {
  	bounds, err := location.NewBoundingBox(
	  	location.Coordinate{Lat: -74.02, Lng: 40.703},
		  location.Coordinate{Lat: -73.96, Lng: 40.8})

    if err != nil {
       log.Fatal(err)

    }
    var historySource location.HistorySource
	  historySource = GetApiHistorySource()
    vis := visualization.NewVisualizer(*imageSize, &historySource, bounds);
    err = vis.GenerateImage("./vis.png");
    if err != nil {
       log.Fatal(err)
    }
  }
}
