package main

import (
	oauth "github.com/hokapoka/goauth"
	"./latitude_api"
	"fmt"
	"log"
)

var googleConsumer *oauth.OAuthConsumer
var accessToken *oauth.AccessToken

func main() {
	googleConsumer = latitude_api.NewConsumer()
	accessToken, err := latitude_api.NewAccessToken(googleConsumer)
	if err != nil{ log.Exit(err) }
	connection := latitude_api.Connection{AccessToken: accessToken, OauthConsumer: googleConsumer}

	locationHistoryUrl := "https://www.googleapis.com/latitude/v1/location"

	params := oauth.Params{
		&oauth.Pair{Key:"key", Value:"AIzaSyDd0W4n2lc03aPFtT0bHJAb2xkNHSduAGE"},
		&oauth.Pair{Key:"granularity", Value:"best"},
		&oauth.Pair{Key:"max-results", Value:"1"},
	}

	body, err := connection.FetchUrl(locationHistoryUrl, params)
	if err != nil{ log.Exit(err) }
	fmt.Println(body)
}
