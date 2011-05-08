package server

import (
	"github.com/mrjones/latvis/latitude"
	"github.com/mrjones/latvis/location"
	"github.com/mrjones/latvis/visualization"
	"github.com/mrjones/oauth"

  "fmt"
  "http"
	"io/ioutil"
	"log"
	"os"
	"rand"
	"strconv"
	"time"
)

var consumer *oauth.Consumer

//todo fix
var requesttokencache map[string]*oauth.RequestToken

func Serve() {
	DoStupidSetup()
  http.HandleFunc("/authorize", Authorize);
  http.HandleFunc("/drawmap", DrawMap);
  http.HandleFunc("/blob", ServeBlob);
  err := http.ListenAndServe(":8081", nil)
  log.Fatal(err)
}

func DoStupidSetup() {
  consumer = latitude.NewConsumer();
	requesttokencache = make(map[string]*oauth.RequestToken)
}

// ======================================
// ============ BLOB STORAGE ============
// ======================================

type Blob struct {
	Data []byte

	// TODO(mrjones): metadata (e.g. Content-Type)
}

type Handle struct {
	timestamp int64
	n1, n2, n3 int64
}

type BlobStore interface {
	// Stores a blob, identified by the Handle, to the BlobStore.
	// Storing a second blob with the same handle will overwrite the first one.
	Store(handle *Handle, blob *Blob) os.Error

	// Fetches the blob with the given handle.
	// TODO(mrjones): distinguish true error from missing blob?
	Fetch(handle *Handle) (*Blob, os.Error)
}

type LocalFSBlobStore struct {
}

func (s *LocalFSBlobStore) Store(handle *Handle, blob *Blob) os.Error {
	filename := s.filename(handle)

	return ioutil.WriteFile(filename, blob.Data, 0600)
}

func (s *LocalFSBlobStore) Fetch(handle *Handle) (*Blob, os.Error) {
	filename := s.filename(handle)
	data, err := ioutil.ReadFile(filename)
	blob := &Blob{Data: data}
	return blob, err
}

func (s *LocalFSBlobStore) filename(h *Handle) string {
	return fmt.Sprintf("images/%d-%d%d%d.png", h.timestamp, h.n1, h.n2, h.n3);
}

// ======================================
// ============ BLOB HELPERS ============
// ======================================

func generateNewHandle() *Handle {
	return &Handle{
		timestamp: time.Seconds(),
		n1: rand.Int63(),
		n2: rand.Int63(),
		n3: rand.Int63(),
	}
}

// TODO(mrjones): generalize
func serializeHandleToUrl(h *Handle) string {
 	return fmt.Sprintf("/blob?s=%d&n1=%d&n2=%d&n3=%d", h.timestamp, h.n1, h.n2, h.n3)
}

func parseHandle(params map[string][]string) (*Handle, os.Error) {
	s, err := extractInt64("s", params)
	if err != nil {
		return nil, err
	}
	n1, err := extractInt64("n1", params)
	if err != nil {
		return nil, err
	}
	n2, err := extractInt64("n2", params)
	if err != nil {
		return nil, err
	}
	n3, err := extractInt64("n3", params)
	if err != nil {
		return nil, err
	}
	return &Handle{timestamp: s, n1: n1, n2: n2, n3: n3}, nil
}

func extractInt64(name string, params map[string][]string) (int64, os.Error) {
	str, err := extractParam(name, params)
	if err != nil {
		return -1, err
	}
	n, err := strconv.Atoi64(str)
	if err != nil {
		return -1, err
	}
	return n, err
}

func extractParam(name string, params map[string][]string) (string, os.Error) {
	if len(params[name]) > 0 {
		return params[name][0], nil
	}
	return "", os.NewError("Missing parameter: '" + name + "'")
}


// ======================================
// ============ SERVER STUFF ============
// ======================================

func serveError(response http.ResponseWriter, err os.Error) {
	serveErrorMessage(response, err.String())
}

func serveErrorMessage(response http.ResponseWriter, message string) {
	response.WriteHeader(http.StatusInternalServerError)
	response.Write([]byte(message))
	response.Flush()
}

func ServeBlob(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	handle, err := parseHandle(request.Form)
	if err != nil {
		serveError(response, err)
	}

	blobstore := LocalFSBlobStore{}

	blob, err := blobstore.Fetch(handle)

	response.SetHeader("Content-Type", "image/png")
	response.Write(blob.Data)
}

func propogateParameter(base string, params map[string][]string, key string) string {
	if len(params[key]) > 0 {
		if len(base) > 0 {
			base = base + "&"
		}
		base = base + key + "=" + params[key][0]
	}
	return base
}

func Authorize(response http.ResponseWriter, request *http.Request) {
  connection := latitude.NewConnectionForConsumer(consumer);

	request.ParseForm()
	latlng := ""
	latlng = propogateParameter(latlng, request.Form, "lllat")
	latlng = propogateParameter(latlng, request.Form, "lllng")
	latlng = propogateParameter(latlng, request.Form, "urlat")
	latlng = propogateParameter(latlng, request.Form, "urlng")
	latlng = propogateParameter(latlng, request.Form, "start")
	latlng = propogateParameter(latlng, request.Form, "end")

  token, url, err := connection.TokenRedirectUrl("http://www.mrjon.es:8081/drawmap?" + latlng)
	requesttokencache[token.Token] = token
  if err != nil {
		serveError(response, err)
  } else {
    http.Redirect(response, request, url, http.StatusFound)
  }
}

func extractCoordinateFromUrl(params map[string][]string, latparam string, lngparam string) (bool, *location.Coordinate, os.Error) {
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

	return false, nil, os.NewError("should never happen")
}

func DrawMap(response http.ResponseWriter, request *http.Request) {
  request.ParseForm()

	found, lowerLeft, err := extractCoordinateFromUrl(request.Form, "lllat", "lllng")
	if !found {
		fmt.Println("Lower Left missing: using default")
		lowerLeft = &location.Coordinate{Lat: 40.703, Lng: -74.02}
	}
	if err != nil {
		serveError(response, err)
		return
	}

	found, upperRight, err := extractCoordinateFromUrl(request.Form, "urlat", "urlng")
	if !found {
		fmt.Println("Upper Right missing: using default")
		upperRight = &location.Coordinate{Lat: 40.8, Lng: -73.96}
	}
	if err != nil {
		serveError(response, err)
		return
	}

	fmt.Printf("Bounding Box: LL[%f,%f], UR[%f,%f]",
		lowerLeft.Lat, lowerLeft.Lng, upperRight.Lat, upperRight.Lng)

	start := &time.Time{Year: 2010, Month: 7, Day: 1}
	if len(request.Form["start"]) > 0 {
		startTs, err := strconv.Atoi64(request.Form["start"][0])
		if err != nil {
			startTs = -1
		}
		start = time.SecondsToUTC(startTs)
	}

	end := &time.Time{Year: 2010, Month: 7, Day: 1}
	if len(request.Form["end"]) > 0 {
		endTs, err := strconv.Atoi64(request.Form["end"][0])
		if err != nil {
			endTs = -1
		}
		end = time.SecondsToUTC(endTs)
	}

	bounds, err := location.NewBoundingBox(*lowerLeft, *upperRight)

	if err != nil {
 		serveError(response, err)
		return
	}

  connection := latitude.NewConnectionForConsumer(consumer)
  if oauthToken, ok := request.Form["oauth_token"]; ok && len(oauthToken) > 0 {
    if oauthVerifier, ok := request.Form["oauth_verifier"]; ok && len(oauthVerifier) > 0 {
			rtoken := requesttokencache[oauthToken[0]]
      atoken, err := connection.ParseToken(rtoken, oauthVerifier[0])
			if err != nil {
 				serveError(response, err)
				return
			}
      var authorizedConnection location.HistorySource
      authorizedConnection = connection.Authorize(atoken)
      vis := visualization.NewVisualizer(512, &authorizedConnection, bounds, *start, *end)
			handle := generateNewHandle()

			data, err := vis.Bytes()
			if err != nil {
 				serveError(response, err)
				return
			}

			store := LocalFSBlobStore{}
			blob := &Blob{Data: *data}
			err = store.Store(handle, blob)
				
			if err != nil {
 				serveError(response, err)
				return
			}

 			url := serializeHandleToUrl(handle)
			http.Redirect(response, request, url, http.StatusFound)
    }
  }
}
