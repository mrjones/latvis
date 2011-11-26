package server

import (
	"github.com/mrjones/latvis/latitude"

	// TODO(mrjones): fix
	"appengine"
	"appengine/taskqueue"

  "fmt"
  "http"
	"log"
	"os"
	"strings"
	"url"
)

var config *ServerConfig

func Setup(serverConfig *ServerConfig) {
	config = serverConfig

  http.HandleFunc("/authorize", AuthorizeHandler)
  http.HandleFunc("/drawmap", DrawMapHandler)
  http.HandleFunc("/async_drawmap", AsyncDrawMapHandler)
  http.HandleFunc("/drawmap_worker", DrawMapWorker)
  http.HandleFunc("/render/", RenderHandler)
	http.HandleFunc("/display/", ResultPageHandler)
	http.HandleFunc("/is_ready/", IsReadyHandler)
//  http.HandleFunc("/async_render/", AsyncRenderHandler)
}

func Serve() {
  err := http.ListenAndServe(":8081", nil)
  log.Fatal(err)
}

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

	blob, err := config.blobStorage.OpenStore(request).Fetch(handle)

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

	// TODO(mrjones): move to an HTML template
	response.Write([]byte("<html><body><div id='canvas' /><img src='/img/spinner.gif' id='spinner' /><br /><div id='debug'/><script type='text/javascript' src='/js/image-loader.js'></script><script type='text/javascript'>loadImage('" + urlParts[2] + "', 5);</script></body></html>"))
}

//func AsyncRenderHandler(response http.ResponseWriter, request *http.Request) {
//	handle, err := parseHandle2(request.URL.Path)
//	if err != nil {
//		serveErrorWithLabel(response, "(Async) parsHandle2 error", err)
//		return
//	}
//	response.Write([]byte(strconv.Itoa64(handle.timestamp)))
//}

func RenderHandler(response http.ResponseWriter, request *http.Request) {
	handle, err := parseHandle2(request.URL.Path)
	if err != nil {
		serveErrorWithLabel(response, "(Sync) parseHandle2 error", err)
		return
	}

	blob, err := config.blobStorage.OpenStore(request).Fetch(handle)

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
	consumer.HttpClient = config.httpClient.GetClient(request)
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

	config.secretStorage.GetStore(request).Store(token.Token, token)
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
  	blobStorage: config.blobStorage,
  	httpClientProvider: config.httpClient,
	  secretStorageProvider: config.secretStorage,
	}

	handle := generateNewHandle();
	err = engine.Render(rr, request, handle)

	if err != nil {
 		serveErrorWithLabel(response, "DrawMapHandler/engine.Render", err)
		return
	}

 	url := serializeHandleToUrl2(handle, "png", "render")
	http.Redirect(response, request, url, http.StatusFound)
}

func AsyncDrawMapHandler(response http.ResponseWriter, request *http.Request) {
  request.ParseForm()

	rr, err := deserializeRenderRequest(&request.Form)
	if err != nil {
 		serveErrorWithLabel(response, "AsyncDrawMapHandler/deserialize", err)
		return
	}

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

 	url := serializeHandleToUrl2(handle, "png", "display")
// 	url := serializeHandleToUrl2(handle, "png", "render")
	http.Redirect(response, request, url, http.StatusFound)
}

func DrawMapWorker(response http.ResponseWriter, request *http.Request) {
	c := appengine.NewContext(request)
	c.Infof("Worker started...") 

	fmt.Println("DrawMapWorker...")
  request.ParseForm()

	rr, err := deserializeRenderRequest(&request.Form)
	if err != nil {
 		serveErrorWithLabel(response, "deserializeRenderRequest() error", err)
		return
	}

	fmt.Printf("DrawMapWorker: start %d -> end %d\n ", rr.start.Seconds(), rr.end.Seconds())

	engine := &RenderEngine{
  	blobStorage: config.blobStorage,
  	httpClientProvider: config.httpClient,
	  secretStorageProvider: config.secretStorage,
	}

	// parse from URL
	handle, err := parseHandleFromParams(&request.Form);
	if err != nil {
 		serveErrorWithLabel(response, "parseHandleFromParams error", err)
		return
	}

	c.Infof("Rendering...") 
	err = engine.Render(rr, request, handle)
	c.Infof("Rendering complete.") 

	if err != nil {
 		serveErrorWithLabel(response, "engine.Render error", err)
		return
	}

	c.Infof("Worker complete.") 
}
