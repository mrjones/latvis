package latvis

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
)

var config *ServerConfig

func UseConfig(serverConfig *ServerConfig) {
	config = serverConfig
}

func Setup(serverConfig *ServerConfig) {
	UseConfig(serverConfig)

	// Starts the process, redirecting to Google for OAuth credentials
	http.HandleFunc("/authorize", AuthorizeHandler)

	// Fetches data and draws a map (synchronously), once the process
	// is complete, it redirects to a page to display the image.
	//
	// NOTE: This endpoint seems like it should be obsolete, and replaced
	//   by "async_drawmap", however, I'm keeping it around for now since
	//   this function is a lot easier to implement on a non-appengine
	//   stack. (I.e. you don't need to supply an implementation of
	//   UrlTaskQueue.)
	http.HandleFunc("/drawmap", SynchronousDrawMapHandler)

	// Asynchronously kicks off a worker to fetch data and generate
	// an image, and then immediately redirects to a page which displays
	// a spinner and polls, waiting for the image to be complete.
	http.HandleFunc("/async_drawmap", AsyncDrawMapHandler)

	// Worker task which fetches the requested data, and renders an image.
	// Writes the result to storage, but doesn't return any data.
	http.HandleFunc("/drawmap_worker", DrawMapWorker)

	// Displays the requested image (as a raw image/png)
	http.HandleFunc("/render/", RenderHandler)

	// Polls, waiting for the requested image to be ready, and once it is
	// displays that image. (This returns text/html, with an embedded <img>
	// referenceing a "/render/" endpoint.)
	http.HandleFunc("/display/", ResultPageHandler)

	// Checks if the requested image is ready or not (used for polling on
	// the "display" page.
	http.HandleFunc("/is_ready/", IsReadyHandler)
}

func Serve() {
	fmt.Println("Localserver Serving on Port 8081")
	err := http.ListenAndServe(":8081", nil)
	log.Fatal(err)
}

func IsReadyHandler(response http.ResponseWriter, request *http.Request) {
	handle, err := parseHandleFromUrl(request.URL.Path)
	if err != nil {
		serveErrorWithLabel(response, "error parsing blob handle", err)
		return
	}

	blob, err := config.blobStorage.OpenStore(request).Fetch(handle)

	if err != nil || blob == nil {
		response.Write([]byte("fail"))
	} else {
		response.Write([]byte("ok"))
	}
}

type ResultPageInfo struct {
	Filename string
}

var resultPageSource = `
<html>
 <head>
  <title>image - latvis.mrjon.es</title>
  <link rel='stylesheet' media='all' href='/css/style.css'>
 </head>
 <body class='latvis-render'>
  <div id='metadata' class='latvis-metadata' style='display:none;'></div>
  <div id='canvas' class='latvis-image'></div>
  <div id='loading' style='width:auto; padding: 5em; text-align: center'>
    <img src='/img/generating.png' id='generating' />
    <br />
    <img src='/img/spinner.gif' id='spinner' />
  </div>
  <br />
  <div id='debug'></div>
  <script type='text/javascript' src='/js/image-loader.js'></script>
  <script type='text/javascript'>loadImage('{{.Filename}}', 5);</script>
  <script type="text/javascript">
   var _gaq = _gaq || [];
   _gaq.push(['_setAccount', 'UA-16767111-2']);
   _gaq.push(['_trackPageview']);
   _gaq.push(['_trackPageLoadTime']);

   (function() {
     var ga = document.createElement('script'); ga.type = 'text/javascript'; ga.async = true;
     ga.src = ('https:' == document.location.protocol ? 'https://ssl' : 'http://www') + '.google-analytics.com/ga.js';
     var s = document.getElementsByTagName('script')[0]; s.parentNode.insertBefore(ga, s);
   })();
  </script>
 </body>
</html>`

func ResultPageHandler(response http.ResponseWriter, request *http.Request) {
	urlParts := strings.Split(request.URL.Path, "/")
	if len(urlParts) != 3 {
		serveError(response, errors.New("Invalid filename [1]: "+request.URL.Path))
	}
	if urlParts[0] != "" {
		serveError(response, errors.New("Invalid filename [2]: "+request.URL.Path))
	}

	t, err := template.New("Result Page").Parse(resultPageSource)
	if err != nil {
		serveErrorWithLabel(response, "Template parsing error", err)
		return
	}
	t.Execute(response, &ResultPageInfo{Filename: urlParts[2]})
}

func RenderHandler(response http.ResponseWriter, request *http.Request) {
	handle, err := parseHandleFromUrl(request.URL.Path)
	if err != nil {
		serveErrorWithLabel(response, "(Sync) parseHandleFromUrl error", err)
		return
	}

	blob, err := config.blobStorage.OpenStore(request).Fetch(handle)

	if err != nil {
		serveErrorWithLabel(response, "RenderHandler/OpenStore error", err)
		return
	}

	if blob == nil {
		serveError(response, errors.New("blob is nil"))
		return
	}

	response.Header().Set("Content-Type", "image/png")
	response.Write(blob.Data)
}

func AuthorizeHandler(response http.ResponseWriter, request *http.Request) {
	connection := config.latitude.NewConnection(request)

	request.ParseForm()
	latlng := ""
	latlng = propogateParameter(latlng, &request.Form, "lllat")
	latlng = propogateParameter(latlng, &request.Form, "lllng")
	latlng = propogateParameter(latlng, &request.Form, "urlat")
	latlng = propogateParameter(latlng, &request.Form, "urlng")
	latlng = propogateParameter(latlng, &request.Form, "start")
	latlng = propogateParameter(latlng, &request.Form, "end")

	protocol := "http"
	if request.TLS != nil {
		protocol = "https"
	}
	redirectUrl := fmt.Sprintf("%s://%s/async_drawmap?%s", protocol, request.Host, latlng)
	//redirectUrl := fmt.Sprintf("%s://%s/drawmap?%s", protocol, request.Host, latlng)

	log.Printf("Redirect URL: '%s'\n", redirectUrl)

	token, url, err := connection.TokenRedirectUrl(redirectUrl)
	if err != nil {
		serveErrorWithLabel(response, "TokenRedirectUrl error", err)
		return
	}

	config.secretStorage.GetStore(request).Store(token.Token, token)
	http.Redirect(response, request, url, http.StatusFound)
}

func SynchronousDrawMapHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	rr, err := deserializeRenderRequest(&request.Form)
	if err != nil {
		serveErrorWithLabel(response, "SynchronousDrawMapHandler/deserializeRenderRequest error", err)
		return
	}

	handle := generateNewHandle()
	err = config.renderEngine.Render(rr, request, handle)

	if err != nil {
		serveErrorWithLabel(response, "SynchronousDrawMapHandler/engine.Render", err)
		return
	}

	url := serializeHandleToUrl(handle, "png", "render")
	http.Redirect(response, request, url, http.StatusFound)
}

func AsyncDrawMapHandler(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	rr, err := deserializeRenderRequest(&request.Form)
	if err != nil {
		serveErrorWithLabel(response, "AsyncDrawMapHandler/deserialize", err)
		return
	}

	handle := generateNewHandle()

	var params = make(url.Values)
	serializeRenderRequest(rr, &params)
	serializeHandleToParams(handle, &params)

	config.taskQueue.GetQueue(request).Enqueue("/drawmap_worker", &params)

	url := serializeHandleToUrl(handle, "png", "display")
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

	fmt.Printf("DrawMapWorker: start %d -> end %d\n ", rr.start.Unix(), rr.end.Unix())

	// parse from URL
	handle, err := parseHandleFromParams(&request.Form)
	if err != nil {
		serveErrorWithLabel(response, "parseHandleFromParams error", err)
		return
	}

	err = config.renderEngine.Render(rr, request, handle)

	if err != nil {
		serveErrorWithLabel(response, "engine.Render error", err)
		return
	}

	response.WriteHeader(http.StatusOK)
}

func serveErrorWithLabel(response http.ResponseWriter, message string, err error) {
	serveErrorMessage(response, message+":"+err.Error())
}

func serveError(response http.ResponseWriter, err error) {
	serveErrorMessage(response, err.Error())
}

func serveErrorMessage(response http.ResponseWriter, message string) {
	fmt.Println("ERROR: " + message)

	response.WriteHeader(http.StatusInternalServerError)
	response.Write([]byte(message))
}