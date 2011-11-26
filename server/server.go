package server

import (
	"github.com/mrjones/latvis/latitude"
	"github.com/mrjones/latvis/location"
	"github.com/mrjones/oauth"

	// TODO(mrjones): fix
	"appengine"
	"appengine/taskqueue"

  "fmt"
  "http"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"url"
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

  http.HandleFunc("/async_drawmap", AsyncDrawMapHandler)
  http.HandleFunc("/drawmap_worker", DrawMapWorker)

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

func extractCoordinateFromUrl(params *url.Values, latparam string, lngparam string) (*location.Coordinate, os.Error) {
	if params.Get(latparam) == "" {
		return nil, os.NewError("Missing required query paramter: " + latparam)
	}
	if params.Get(lngparam) == "" {
		return nil, os.NewError("Missing required query paramter: " + lngparam)
	}

	lat, err := strconv.Atof64(params.Get(latparam))
	if err != nil {
		return nil, err
	}
	lng, err := strconv.Atof64(params.Get(lngparam))
	if err != nil {
		return nil, err
	}
	
	return &location.Coordinate{Lat: lat, Lng: lng}, nil
}


func extractTimeFromUrl(params *url.Values, param string) (*time.Time, os.Error) {
	if params.Get(param) == "" {
		return nil, os.NewError("Missing query param: " + param)
	}
	startTs, err := strconv.Atoi64(params.Get(param))
	if err != nil {
		startTs = -1
	}
	return time.SecondsToUTC(startTs), nil
}

func extractStringFromUrl(params *url.Values, param string) (string, os.Error) {
	if params.Get(param) == "" {
		return "", os.NewError("Missing query param: " + param)
	}
	return params.Get(param), nil
}

func propogateParameter(base string, params *url.Values, key string) string {
	if params.Get(key) != "" {
		if len(base) > 0 {
			base = base + "&"
		}
		// TODO(mrjones): sigh use the right library
		base = base + key + "=" + url.QueryEscape(params.Get(key))
	}
	return base
}

// ======================================
// ============ SERVER STUFF ============
// ======================================

func serveErrorWithLabel(response http.ResponseWriter, message string, err os.Error) {
	serveErrorMessage(response, message + ":" + err.String())
}

func serveError(response http.ResponseWriter, err os.Error) {
	serveErrorMessage(response, err.String())
}

func serveErrorMessage(response http.ResponseWriter, message string) {
	fmt.Println("ERROR: " + message)

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
		serveErrorWithLabel(response, "(Async) parsHandle2 error", err)
		return
	}
	response.Write([]byte(strconv.Itoa64(handle.timestamp)))
}

func RenderHandler(response http.ResponseWriter, request *http.Request) {
	handle, err := parseHandle2(request.URL.Path)
	if err != nil {
		serveErrorWithLabel(response, "(Sync) parseHandle2 error", err)
		return
	}

	blob, err := storage.OpenStore(request).Fetch(handle)

	if err != nil {
		serveErrorWithLabel(response, "RenderHandler/OpenStore error", err)
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
	latlng = propogateParameter(latlng, &request.Form, "lllat")
	latlng = propogateParameter(latlng, &request.Form, "lllng")
	latlng = propogateParameter(latlng, &request.Form, "urlat")
	latlng = propogateParameter(latlng, &request.Form, "urlng")
	latlng = propogateParameter(latlng, &request.Form, "start")
	latlng = propogateParameter(latlng, &request.Form, "end")

	protocol := "http"
	if (request.TLS != nil) {
		protocol = "https"
	}
	redirectUrl := fmt.Sprintf("%s://%s/async_drawmap?%s", protocol, request.Host, latlng)
//	redirectUrl := fmt.Sprintf("%s://%s/drawmap?%s", protocol, request.Host, latlng)

	log.Printf("Redirect URL: '%s'\n", redirectUrl)

  token, url, err := connection.TokenRedirectUrl(redirectUrl)
	if err != nil {
 		serveErrorWithLabel(response, "TokenRedirectUrl error", err)
		return
	}

	requesttokencache[token.Token] = token
	secretStoreProvider.GetStore(request).Store(token.Token, token)
  http.Redirect(response, request, url, http.StatusFound)
}

func DrawMapHandler(response http.ResponseWriter, request *http.Request) {
  request.ParseForm()

	rr, err := deserializeRenderRequest(&request.Form)
	if err != nil {
 		serveErrorWithLabel(response, "DrawMapHandler/deserializeRenderRequest error", err)
		return
	}

	engine := &RenderEngine{
  	httpClientProvider: clientProvider,
	  secretStorageProvider: secretStoreProvider,
	}

	handle := generateNewHandle();
	err = engine.Render(rr, request, handle)

	if err != nil {
 		serveErrorWithLabel(response, "DrawMapHandler/engine.Render", err)
		return
	}

 	url := serializeHandleToUrl2(handle, "png", "")
	http.Redirect(response, request, url, http.StatusFound)
}

func AsyncDrawMapHandler(response http.ResponseWriter, request *http.Request) {
  request.ParseForm()

	rr, err := deserializeRenderRequest(&request.Form)
	if err != nil {
 		serveErrorWithLabel(response, "AsyncDrawMapHandler/deserialize", err)
		return
	}

	fmt.Printf("AsyncDrawMapHandler: start %d -> end %d\n ", rr.start.Seconds(), rr.end.Seconds())

	handle := generateNewHandle();

	c := appengine.NewContext(request)

	var params = make(url.Values)
	serializeRenderRequest(rr, &params)
	serializeHandleToParams(handle, &params)

	t := taskqueue.NewPOSTTask("/drawmap_worker", params)
  if _, err := taskqueue.Add(c, t, ""); err != nil {
		http.Error(response, err.String(), http.StatusInternalServerError)
		return
	}

 	url := serializeHandleToUrl2(handle, "png", "async_")
	http.Redirect(response, request, url, http.StatusFound)
}

func DrawMapWorker(response http.ResponseWriter, request *http.Request) {
	fmt.Println("DrawMapWorker...")
  request.ParseForm()

	rr, err := deserializeRenderRequest(&request.Form)
	if err != nil {
 		serveErrorWithLabel(response, "deserializeRenderRequest() error", err)
		return
	}

	fmt.Printf("DrawMapWorker: start %d -> end %d\n ", rr.start.Seconds(), rr.end.Seconds())

	engine := &RenderEngine{
  	httpClientProvider: clientProvider,
	  secretStorageProvider: secretStoreProvider,
	}

	// parse from URL
	handle, err := parseHandleFromParams(&request.Form);
	if err != nil {
 		serveErrorWithLabel(response, "parseHandleFromParams error", err)
		return
	}

	err = engine.Render(rr, request, handle)

	if err != nil {
 		serveErrorWithLabel(response, "engine.Render error", err)
		return
	}

	// update storage
}

