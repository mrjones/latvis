package main

import (
	oauth "github.com/hokapoka/goauth"
	"./latitude_api"
	"fmt"
	"log"
)

var accessToken *oauth.AccessToken

func main() {
	connection := latitude_api.NewConnection()
	accessToken, err := connection.NewAccessToken()
	if err != nil{ log.Exit(err) }
	authConnection := connection.Authorize(accessToken);

	locationHistoryUrl := "https://www.googleapis.com/latitude/v1/location"

	params := oauth.Params{
		&oauth.Pair{Key:"granularity", Value:"best"},
		&oauth.Pair{Key:"max-results", Value:"1"},
	}

	body, err := authConnection.FetchUrl(locationHistoryUrl, params)
	if err != nil{ log.Exit(err) }
	fmt.Println(body)
}
