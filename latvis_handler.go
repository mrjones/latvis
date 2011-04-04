package latvis_handler

import (
  "fmt"
  "http"
  "./latitude_api"
  "./location"
	oauth "github.com/hokapoka/goauth"
  "./visualizer"
)

var consumer *oauth.OAuthConsumer

func DoStupidSetup() {
  consumer = latitude_api.NewConsumer("http://www.mrjon.es:8080/drawmap");
}

func ServePng(response http.ResponseWriter, request *http.Request) {
  http.ServeFile(response, request, "vis-web.png")
}

func Authorize(response http.ResponseWriter, request *http.Request) {
  connection := latitude_api.NewConnectionForConsumer(consumer);
  url, err := connection.TokenRedirectUrl()
  if err != nil {
    fmt.Fprintf(response, err.String())
  } else {
    http.Redirect(response, request, *url, http.StatusFound)
  }
}

func DrawMap(response http.ResponseWriter, request *http.Request) {
  connection := latitude_api.NewConnectionForConsumer(consumer)
  request.ParseForm()
  if oauthToken, ok := request.Form["oauth_token"]; ok && len(oauthToken) > 0 {
    if oauthVerifier, ok := request.Form["oauth_verifier"]; ok && len(oauthVerifier) > 0 {
      token := connection.ParseToken(oauthToken[0], oauthVerifier[0])
      var authorizedConnection location.HistorySource
      authorizedConnection = connection.Authorize(token)
      vis := visualizer.NewVisualizer(512, &authorizedConnection)
      vis.GenerateImage("vis-web.png")
      http.Redirect(response, request, "/latestimage", http.StatusFound)
    }
  }
}
