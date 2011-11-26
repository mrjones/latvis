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
	"strings"
	"time"
)

//var consumer *oauth.Consumer
var storage HttpBlobStoreProvider
var clientProvider HttpClientProvider
var secretStoreProvider HttpOauthSecretStoreProvider

//todo fix
var requesttokencache map[string]*oauth.RequestToken

func Setup(blobStoreProvider HttpBlobStoreProvider, httpClientProvider HttpClientProvider) {
	DoStupidSetup()
	storage = blobStoreProvider
	clientProvider = httpClientProvider
	secretStoreProvider = &InMemoryOauthSecretStoreProvider{}

  http.HandleFunc("/authorize", AuthorizeHandler)
  http.HandleFunc("/drawmap", DrawMapHandler)
  http.HandleFunc("/render/", RenderHandler)

	http.HandleFunc("/display/", ResultPageHandler)
	http.HandleFunc("/is_ready/", IsReadyHandler)
  http.HandleFunc("/async_render/", AsyncRenderHandler)
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

type HttpOauthSecretStoreProvider interface {
	GetStore(req *http.Request) OauthSecretStore
}

type OauthSecretStore interface {
	Store(tokenString string, token *oauth.RequestToken)
	Lookup(tokenString string) *oauth.RequestToken
}

//

type InMemoryOauthSecretStoreProvider struct {
	storage *InMemoryOauthSecretStore
}

func (p *InMemoryOauthSecretStoreProvider) GetStore(req *http.Request) OauthSecretStore {
	if (p.storage == nil) {
		// todo threads
		p.storage = NewInMemoryOauthSecretStore()
	}
	return p.storage
}

type InMemoryOauthSecretStore struct {
	store map[string]*oauth.RequestToken
}

func NewInMemoryOauthSecretStore() *InMemoryOauthSecretStore {
	return &InMemoryOauthSecretStore{
	  store: make(map[string]*oauth.RequestToken),
	}
}

func (s *InMemoryOauthSecretStore) Store(tokenString string, token *oauth.RequestToken) {
	s.store[tokenString] = token
}

func (s *InMemoryOauthSecretStore) Lookup(tokenString string) *oauth.RequestToken {
	return s.store[tokenString]
}

//

type StandardHttpClientProvider struct {
}

func (s *StandardHttpClientProvider) GetClient(req *http.Request) oauth.HttpClient{
	return &http.Client{}
}

//

type LocalFSBlobStoreProvider struct {
}

func (p *LocalFSBlobStoreProvider) OpenStore(req *http.Request) BlobStore {
	return &LocalFSBlobStore{}
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

func IsReadyHandler(response http.ResponseWriter, request *http.Request) {
	handle, err := parseHandle2(request.URL.Path)
	if err != nil {
		response.Write([]byte("error: " + err.String()))
		return
	}

	blob, err := storage.OpenStore(request).Fetch(handle)

	if err != nil || blob == nil {
		response.Write([]byte("fail"))
	} else {
		response.Write([]byte("ok"))
	}
}

func ResultPageHandler(response http.ResponseWriter, request *http.Request) {
	urlParts := strings.Split(request.URL.Path, "/")
	if len(urlParts) != 3 {
		serveError(response, os.NewError("Invalid filename [1]: " + request.URL.Path))
	}
	if urlParts[0] != "" {
		serveError(response, os.NewError("Invalid filename [2]: " + request.URL.Path))
	}

	response.Write([]byte("<html><body><div id='canvas' /><img src='/img/spinner.gif' id='spinner' /><br /><div id='debug'/><script type='text/javascript' src='/js/image-loader.js'></script><script type='text/javascript'>loadImage('" + urlParts[2] + "', 5);</script></body></html>"))
}

func AsyncRenderHandler(response http.ResponseWriter, request *http.Request) {
	handle, err := parseHandle2(request.URL.Path)
	if err != nil {
		serveError(response, err)
		return
	}
	response.Write([]byte(strconv.Itoa64(handle.timestamp)))
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
	secretStoreProvider.GetStore(request).Store(token.Token, token)
  http.Redirect(response, request, url, http.StatusFound)
}

func Render(renderRequest *RenderRequest, httpRequest *http.Request) (*Handle, os.Error) {
  consumer := latitude.NewConsumer();
	consumer.HttpClient = clientProvider.GetClient(httpRequest)
  connection := latitude.NewConnectionForConsumer(consumer)

	rtoken := secretStoreProvider.GetStore(httpRequest).Lookup(renderRequest.oauthToken);
//	rtoken := &oauth.RequestToken{
//    Token: renderRequest.oauthToken,
//    Secret: secret,
//	}
//	rtoken := requesttokencache[renderRequest.oauthToken]
	if (rtoken == nil) {
		return nil, os.NewError("No token stored for: " + renderRequest.oauthToken)
	}
  atoken, err := connection.ParseToken(rtoken, renderRequest.oauthVerifier)

	if err != nil { return nil, err }
  
	var authorizedConnection location.HistorySource
  authorizedConnection = connection.Authorize(atoken)
  vis := visualization.NewVisualizer(512, &authorizedConnection, renderRequest.bounds, *renderRequest.start, *renderRequest.end)

	data, err := vis.Bytes()
	if err != nil { return nil, err }

	handle := generateNewHandle()
	blob := &Blob{Data: *data}
	err = storage.OpenStore(httpRequest).Store(handle, blob)
	if err != nil { return nil, err }

	return handle, nil
}

func DrawMapHandler(response http.ResponseWriter, request *http.Request) {
  request.ParseForm()

	rr, err := parseRenderRequest(request.Form)
	if err != nil {
 		serveError(response, err)
		return
	}

	handle, err := Render(rr, request)

	if err != nil {
 		serveError(response, err)
		return
	}

 	url := serializeHandleToUrl2(handle, "png")
// 	url := serializeHandleToUrl(handle)
	http.Redirect(response, request, url, http.StatusFound)
}
