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
	"strings"
	"time"
)

//var consumer *oauth.Consumer
var storage HttpBlobStoreProvider
var clientProvider HttpClientProvider

//todo fix
var requesttokencache map[string]*oauth.RequestToken

func Setup(blobStoreProvider HttpBlobStoreProvider, httpClientProvider HttpClientProvider) {
	DoStupidSetup()
	storage = blobStoreProvider
	clientProvider = httpClientProvider
  http.HandleFunc("/authorize", AuthorizeHandler)
  http.HandleFunc("/drawmap", DrawMapHandler)
  http.HandleFunc("/render/", RenderHandler)
}

func Serve() {
  err := http.ListenAndServe(":8081", nil)
  log.Fatal(err)
}

func DoStupidSetup() {
//  consumer = latitude.NewConsumer();
	requesttokencache = make(map[string]*oauth.RequestToken)
}

// Appengine hacks:
// Using appengine services (datastore, urlfetcher) need an appengine.Context
// which requires the http.Request at construction time.
// These interfaces are for isolating the appengine specific code, but are still
// awkward since they require an http.Request to construct seemingly unrelated
// objects.
type HttpBlobStoreProvider interface {
	OpenStore(req *http.Request) BlobStore
}

type HttpClientProvider interface {
	GetClient(req *http.Request) oauth.HttpClient
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

func (h *Handle) String() string {
	return fmt.Sprintf("%d-%d%d%d", h.timestamp, h.n1, h.n2, h.n3)
}

type BlobStore interface {
	// Stores a blob, identified by the Handle, to the BlobStore.
	// Storing a second blob with the same handle will overwrite the first one.
	Store(handle *Handle, blob *Blob) os.Error

	// Fetches the blob with the given handle.
	// TODO(mrjones): distinguish true error from missing blob?
	Fetch(handle *Handle) (*Blob, os.Error)
}

type StandardHttpClientProvider struct {
}

func (s *StandardHttpClientProvider) GetClient(req *http.Request) oauth.HttpClient{
	return &http.Client{}
}

type LocalFSBlobStoreProvider struct {
}

func (p *LocalFSBlobStoreProvider) OpenStore(req *http.Request) BlobStore {
	return &LocalFSBlobStore{}
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

func serializeHandleToUrl2(h *Handle, suffix string) string {
 	return fmt.Sprintf("/render/%d-%d-%d-%d.%s", h.timestamp, h.n1, h.n2, h.n3, suffix)
}

func parseHandle2(fullpath string) (*Handle, os.Error) {
	directories := strings.Split(fullpath, "/", -1)
	if len(directories) != 3 {
		return nil, os.NewError("Invalid filename [1]: " + fullpath)
	}
	if directories[0] != "" {
		return nil, os.NewError("Invalid filename [2]: " + fullpath)
	}

	filename := directories[2]
	fileparts := strings.Split(filename, ".", -1)

	if len(fileparts) != 2 {
		return nil, os.NewError("Invalid filename [3]: " + fullpath)
	}


	pieces := strings.Split(fileparts[0], "-", -1)
	if len(pieces) != 4 {
		return nil, os.NewError("Invalid filename [4]: " + fullpath)
	}


	s, err := strconv.Atoi64(pieces[0])
	if err != nil {
		return nil, err
	}
	n1, err := strconv.Atoi64(pieces[1])
	if err != nil {
		return nil, err
	}
	n2, err := strconv.Atoi64(pieces[2])
	if err != nil {
		return nil, err
	}
	n3, err := strconv.Atoi64(pieces[3])
	if err != nil {
		return nil, err
	}
	return &Handle{timestamp: s, n1: n1, n2: n2, n3: n3}, nil
}


// ======================================
// ============ URL PARSING =============
// ======================================

func extractCoordinateFromUrl(params map[string][]string, latparam string, lngparam string) (*location.Coordinate, os.Error) {
	if len(params[latparam]) == 0 {
		return nil, os.NewError("Missing required query paramter: " + latparam)
	}
	if len(params[lngparam]) == 0 {
		return nil, os.NewError("Missing required query paramter: " + lngparam)
	}

	lat, err := strconv.Atof64(params[latparam][0])
	if err != nil {
		return nil, err
	}
	lng, err := strconv.Atof64(params[lngparam][0])
	if err != nil {
		return nil, err
	}
	
	return &location.Coordinate{Lat: lat, Lng: lng}, nil
}


func extractTimeFromUrl(params map[string][]string, param string) (*time.Time, os.Error) {
	if len(params[param]) < 1 {
		return nil, os.NewError("Missing query param: " + param)
	}
	startTs, err := strconv.Atoi64(params[param][0])
	if err != nil {
		startTs = -1
	}
	return time.SecondsToUTC(startTs), nil
}

func extractStringFromUrl(params map[string][]string, param string) (string, os.Error) {
	if len(params[param]) < 1 {
		return "", os.NewError("Missing query param: " + param)
	}
	return params[param][0], nil
}

type RenderRequest struct {
	bounds *location.BoundingBox
	start, end *time.Time
	oauthToken string
	oauthVerifier string
}

func parseRenderRequest(params map[string][]string) (*RenderRequest, os.Error) {
	// Parse all input parameters from the URL
	lowerLeft, err := extractCoordinateFromUrl(params, "lllat", "lllng")
	if err != nil {
		return nil, err
	}

	upperRight, err := extractCoordinateFromUrl(params, "urlat", "urlng")
	if err != nil {
		return nil, err
	}

	fmt.Printf("Bounding Box: LL[%f,%f], UR[%f,%f]",
		lowerLeft.Lat, lowerLeft.Lng, upperRight.Lat, upperRight.Lng)

	start, err := extractTimeFromUrl(params, "start")
	if err != nil {
		return nil, err
	}

	end, err := extractTimeFromUrl(params, "end")
	if err != nil {
		return nil, err
	}

	bounds, err := location.NewBoundingBox(*lowerLeft, *upperRight)
	if err != nil {
		return nil, err
	}

	oauthToken, err := extractStringFromUrl(params, "oauth_token")
	if err != nil {
		return nil, err
	}

	oauthVerifier, err := extractStringFromUrl(params, "oauth_verifier")
	if err != nil {
		return nil, err
	}

	return &RenderRequest{bounds: bounds, start: start, end:end, oauthToken: oauthToken, oauthVerifier: oauthVerifier}, nil
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

// ======================================
// ============ SERVER STUFF ============
// ======================================

func serveError(response http.ResponseWriter, err os.Error) {
	serveErrorMessage(response, err.String())
}

func serveErrorMessage(response http.ResponseWriter, message string) {
	response.WriteHeader(http.StatusInternalServerError)
	response.Write([]byte(message))
}

func RenderHandler(response http.ResponseWriter, request *http.Request) {
	handle, err := parseHandle2(request.URL.Path)
	if err != nil {
		serveError(response, err)
		return
	}

	blob, err := storage.OpenStore(request).Fetch(handle)

	if err != nil {
		serveError(response, err)
		return
	}

	if blob == nil {
		serveError(response, os.NewError("blob is nil"))
		return
	}

	response.Header().Set("Content-Type", "image/png")
	response.Write(blob.Data)
}

func AuthorizeHandler(response http.ResponseWriter, request *http.Request) {
  consumer := latitude.NewConsumer();
	consumer.HttpClient = clientProvider.GetClient(request)
  connection := latitude.NewConnectionForConsumer(consumer);

	request.ParseForm()
	latlng := ""
	latlng = propogateParameter(latlng, request.Form, "lllat")
	latlng = propogateParameter(latlng, request.Form, "lllng")
	latlng = propogateParameter(latlng, request.Form, "urlat")
	latlng = propogateParameter(latlng, request.Form, "urlng")
	latlng = propogateParameter(latlng, request.Form, "start")
	latlng = propogateParameter(latlng, request.Form, "end")

	protocol := "http"
	if (request.TLS != nil) {
		protocol = "https"
	}
	redirectUrl := fmt.Sprintf("%s://%s/drawmap?%s", protocol, request.Host, latlng)

	log.Printf("Redirect URL: '%s'\n", redirectUrl)

  token, url, err := connection.TokenRedirectUrl(redirectUrl)
	if err != nil {
 		serveError(response, err)
		return
	}

	requesttokencache[token.Token] = token
  http.Redirect(response, request, url, http.StatusFound)
}

func DrawMapHandler(response http.ResponseWriter, request *http.Request) {
  request.ParseForm()

	rr, err := parseRenderRequest(request.Form)
	if err != nil {
 		serveError(response, err)
		return
	}

  consumer := latitude.NewConsumer();
	consumer.HttpClient = clientProvider.GetClient(request)
  connection := latitude.NewConnectionForConsumer(consumer)
	rtoken := requesttokencache[rr.oauthToken]
  atoken, err := connection.ParseToken(rtoken, rr.oauthVerifier)
	
	if err != nil {
 		serveError(response, err)
		return
	}
  
	var authorizedConnection location.HistorySource
  authorizedConnection = connection.Authorize(atoken)
  vis := visualization.NewVisualizer(512, &authorizedConnection, rr.bounds, *rr.start, *rr.end)

	data, err := vis.Bytes()
	if err != nil {
 		serveError(response, err)
		return
	}

	handle := generateNewHandle()
	blob := &Blob{Data: *data}
	err = storage.OpenStore(request).Store(handle, blob)
				
	if err != nil {
 		serveError(response, err)
		return
	}

 	url := serializeHandleToUrl2(handle, "png")
// 	url := serializeHandleToUrl(handle)
	http.Redirect(response, request, url, http.StatusFound)
}
