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

var envFactory EnvironmentFactory

func UseEnvironmentFactory(serverEnvFactory EnvironmentFactory) {
	envFactory = serverEnvFactory
}

func Setup(serverEnvFactory EnvironmentFactory) {
	UseEnvironmentFactory(serverEnvFactory)

	// Starts the process, redirecting to Google for OAuth credentials
	http.HandleFunc("/authorize", AuthorizeHandler)

	// Asynchronously kicks off a worker to fetch data and generate
	// an image, and then immediately redirects to a page which displays
	// a spinner and polls, waiting for the image to be complete.
	http.HandleFunc("/async_drawmap", AsyncDrawMapHandler)

	// Worker task which fetches the requested data, and renders an image.
	// Writes the result to storage, but doesn't return any data.
	http.HandleFunc("/drawmap_worker", DrawMapWorker)

	// Displays the requested image (as a raw image/png)
	// NOTE: also update static/js/image-loader.js.
	http.HandleFunc("/rawimg/", RenderHandler)

	// Polls, waiting for the requested image to be ready, and once it is
	// displays that image. (This returns text/html, with an embedded <img>
	// referenceing a "/render/" endpoint.)
	http.HandleFunc("/display/", ResultPageHandler)

	// Checks if the requested image is ready or not (used for polling on
	// the "display" page.
	http.HandleFunc("/is_ready/", IsReadyHandler)

	http.Handle("/", http.FileServer(http.Dir("static")))
}

func Serve() {
	fmt.Println("Localserver Serving on Port 8081")
	err := http.ListenAndServe(":8081", nil)
	log.Fatal(err)
}

func IsReadyHandler(response http.ResponseWriter, request *http.Request) {
	env := envFactory.ForRequest(request)
	handle, err := parseHandleFromUrl(request.URL.Path)
	if err != nil {
		serveErrorWithLabel(response, "error parsing blob handle", err)
		return
	}

	blob, err := env.RenderEngineForRequest(request).FetchImage(handle)

	if err != nil || blob == nil {
		log.Println(err)
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
	env := envFactory.ForRequest(request)
	handle, err := parseHandleFromUrl(request.URL.Path)
	if err != nil {
		serveErrorWithLabel(response, "(Sync) parseHandleFromUrl error", err)
		return
	}

	blob, err := env.RenderEngineForRequest(request).FetchImage(handle)
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

func propogateParameter(base string, params *url.Values, key string) string {
	if params.Get(key) != "" {
		if len(base) > 0 {
			base = base + "&"
		}
		// TODO(mrjones): sigh use the right library
		//		base = base + key + "=" + url.QueryEscape(params.Get(key))
		base = base + key + "=" + params.Get(key)
	}
	return base
}

func callbackUrlFor(request *http.Request) string {
	protocol := "http"
	if request.TLS != nil {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/async_drawmap", protocol, request.Host)

}

func AuthorizeHandler(response http.ResponseWriter, request *http.Request) {
	env := envFactory.ForRequest(request)
	request.ParseForm()
	state := ""
	state = propogateParameter(state, &request.Form, "lllat")
	state = propogateParameter(state, &request.Form, "lllng")
	state = propogateParameter(state, &request.Form, "urlat")
	state = propogateParameter(state, &request.Form, "urlng")
	state = propogateParameter(state, &request.Form, "start")
	state = propogateParameter(state, &request.Form, "end")

	callbackUrl := callbackUrlFor(request)
	log.Printf("Callback URL: '%s' + '%s'\n", callbackUrl, state)

	authUrl := env.RenderEngineForRequest(request).GetOAuthUrl(callbackUrl, state)
	http.Redirect(response, request, authUrl, http.StatusFound)
}

func AsyncDrawMapHandler(response http.ResponseWriter, request *http.Request) {	
	env := envFactory.ForRequest(request)

	fmt.Println("--> AsyncDrawMapHandler: " + request.Host + " / " + request.RequestURI)
	request.ParseForm()

	rr, err := deserializeRenderRequest(&request.Form)
	if err != nil {
		serveErrorWithLabel(response, "AsyncDrawMapHandler/deserialize", err)
		return
	}

	handle := GenerateHandle()

	var params = make(url.Values)
	serializeRenderRequest(rr, &params)
	serializeHandleToParams(handle, &params)
	params.Set("verification_code", request.Form.Get("code"))

	env.taskQueue.Enqueue("/drawmap_worker", &params)

	displayImageUrl := serializeHandleToUrl(handle, "png", "display")
	http.Redirect(response, request, displayImageUrl, http.StatusFound)
}

func DrawMapWorker(response http.ResponseWriter, request *http.Request) {
	env := envFactory.ForRequest(request)
	fmt.Println("--> DrawMapWorker: " + request.Host + " / " + request.RequestURI)
	request.ParseForm()

	rr, err := deserializeRenderRequest(&request.Form)
	if err != nil {
    env.Errorf("deserializeRenderRequest: %s", err)
		serveErrorWithLabel(response, "deserializeRenderRequest() error", err)
		return
	}

	handle, err := parseHandleFromParams(&request.Form)
	if err != nil {
    env.Errorf("parseHandleFromParams: %s", err)
		serveErrorWithLabel(response, "parseHandleFromParams error", err)
		return
	}

	verificationCode := request.FormValue("verification_code")
	if verificationCode == "" {
    env.Errorf("verification_code query parameter missing")
		serveErrorWithLabel(response, "get verificationcode", errors.New("verification_code query parameter missing"))
	}

	callbackUrl := callbackUrlFor(request)
	log.Printf("Callback URL: '%s'\n", callbackUrl)
	err = env.RenderEngineForRequest(request).Execute(rr, verificationCode, callbackUrl, handle)
	if err != nil {
    env.Errorf("renderEngine error: %s", err)
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
