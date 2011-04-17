package server

import (
	"github.com/mrjones/latvis/latitude"
	"github.com/mrjones/latvis/location"
	"github.com/mrjones/latvis/visualization"
	"github.com/mrjones/oauth"

  "fmt"
  "http"
	"log"
	"os"
	"strconv"
)

var consumer *oauth.Consumer

//todo fix
var requesttokencache map[string]*oauth.RequestToken

func Serve() {
	DoStupidSetup()
  http.HandleFunc("/authorize", Authorize);
  http.HandleFunc("/drawmap", DrawMap);
  http.HandleFunc("/latestimage", ServePng);
  err := http.ListenAndServe(":8081", nil)
  log.Fatal(err)
}

func DoStupidSetup() {
  consumer = latitude.NewConsumer();
	requesttokencache = make(map[string]*oauth.RequestToken)
}

func ServePng(response http.ResponseWriter, request *http.Request) {
  http.ServeFile(response, request, "vis-web.png")
}

func Authorize(response http.ResponseWriter, request *http.Request) {
  connection := latitude.NewConnectionForConsumer(consumer);

  token, url, err := connection.TokenRedirectUrl("http://www.mrjon.es:8081/drawmap")
//  token, url, err := connection.TokenRedirectUrl("http://www.mrjon.es:8081/drawmap?lllat=37.416936&lllng=-122.092438&urlat=37.423753&urlng=-122.076130")
//  token, url, err := connection.TokenRedirectUrl("http://www.mrjon.es:8081/drawmap?lllat=40.699902&lllng=-74.020386&urlat=40.719811&urlng=-73.970604")
	requesttokencache[token.Token] = token
  if err != nil {
    fmt.Fprintf(response, err.String())
  } else {
    http.Redirect(response, request, url, http.StatusFound)
  }
}

func extractCoordinate(params map[string][]string, latparam string, lngparam string) (bool, *location.Coordinate, os.Error) {
	if len(params[latparam]) > 0 && len(params[lngparam]) > 0 {
		lat, laterr := strconv.Atof64(params[latparam][0])
		lng, lngerr := strconv.Atof64(params[lngparam][0])
		if lngerr == nil && laterr == nil {			
			return true, &location.Coordinate{Lat: lat, Lng: lng}, nil
		} else if laterr != nil {
			fmt.Println("laterr " + laterr.String())
			return false, nil, laterr
		} else if lngerr != nil {
			fmt.Println("lngerr " + laterr.String())
			return false, nil, lngerr
		}
	}

	return false, nil, nil
}

func DrawMap(response http.ResponseWriter, request *http.Request) {
  request.ParseForm()

	found, lowerLeft, err := extractCoordinate(request.Form, "lllat", "lllng")
	if !found {
		fmt.Println("Lower Left missing: using default")
		lowerLeft = &location.Coordinate{Lat: 40.703, Lng: -74.02}
	}
	if err != nil {
		// todo don'tcrash
		log.Fatal(err)
	}

	found, upperRight, err := extractCoordinate(request.Form, "urlat", "urlng")
	if !found {
		fmt.Println("Upper Right missing: using default")
		upperRight = &location.Coordinate{Lat: 40.8, Lng: -73.96}
	}
	if err != nil {
		// todo don'tcrash
		log.Fatal(err)
	}

	fmt.Printf("Bounding Box: LL[%f,%f], UR[%f,%f]",
		lowerLeft.Lat, lowerLeft.Lng, upperRight.Lat, upperRight.Lng)

	bounds, err := location.NewBoundingBox(*lowerLeft, *upperRight)

	if err != nil {
		log.Fatal(err)
//		response.WriteHeader(http.StatusInternalServerError)
//		response.Write([]byte(err.String()))
//		response.Flush()
	}

  connection := latitude.NewConnectionForConsumer(consumer)
  if oauthToken, ok := request.Form["oauth_token"]; ok && len(oauthToken) > 0 {
    if oauthVerifier, ok := request.Form["oauth_verifier"]; ok && len(oauthVerifier) > 0 {
			rtoken := requesttokencache[oauthToken[0]]
      atoken, err := connection.ParseToken(rtoken, oauthVerifier[0])
			if err != nil {
				log.Fatal(err)
			}
      var authorizedConnection location.HistorySource
      authorizedConnection = connection.Authorize(atoken)
      vis := visualization.NewVisualizer(512, &authorizedConnection, bounds)
      err = vis.GenerateImage("vis-web.png")
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(err.String()))
				response.Flush()
			} else {
				http.Redirect(response, request, "/latestimage", http.StatusFound)
			}
    }
  }
}
