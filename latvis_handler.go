package latvis_handler

import (
	"github.com/mrjones/latvis/latitude"
	"github.com/mrjones/latvis/location"
	"github.com/mrjones/latvis/visualization"
	"github.com/mrjones/oauth"

  "fmt"
  "http"
	"log"
)

var consumer *oauth.Consumer

//todo fix
var requesttokencache map[string]*oauth.RequestToken

func DoStupidSetup() {
  consumer = latitude.NewConsumer("http://www.mrjon.es:8081/drawmap");
	requesttokencache = make(map[string]*oauth.RequestToken)
}

func ServePng(response http.ResponseWriter, request *http.Request) {
  http.ServeFile(response, request, "vis-web.png")
}

func Authorize(response http.ResponseWriter, request *http.Request) {
  connection := latitude.NewConnectionForConsumer(consumer);
  token, url, err := connection.TokenRedirectUrl()
	requesttokencache[token.Token] = token
  if err != nil {
    fmt.Fprintf(response, err.String())
  } else {
    http.Redirect(response, request, url, http.StatusFound)
  }
}

func DrawMap(response http.ResponseWriter, request *http.Request) {
  connection := latitude.NewConnectionForConsumer(consumer)
  request.ParseForm()
  if oauthToken, ok := request.Form["oauth_token"]; ok && len(oauthToken) > 0 {
    if oauthVerifier, ok := request.Form["oauth_verifier"]; ok && len(oauthVerifier) > 0 {
			rtoken := requesttokencache[oauthToken[0]]
      atoken, err := connection.ParseToken(rtoken, oauthVerifier[0])
			if err != nil {
				log.Fatal(err)
			}
      var authorizedConnection location.HistorySource
      authorizedConnection = connection.Authorize(atoken)
      vis := visualization.NewVisualizer(512, &authorizedConnection)
      vis.GenerateImage("vis-web.png")
      http.Redirect(response, request, "/latestimage", http.StatusFound)
    }
  }
}
