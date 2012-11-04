package latvis

import (
	"log"
	"net/http"
	"net/url"
)

// ======================================
// ========== ENVIRONMENT API ===========
// ======================================

// Environment encapsulates the dependencies for a latvis server.
//
// Applications should prefer to access system level services (data storage, network
// requests, etc.) via the Environment rather than directly in order to support
// both unit-testing, and also portability (e.g. to the Google Appengine sandbox).
type Environment struct {
	blobStore        BlobStore
	taskQueue        UrlTaskQueue
	mockRenderEngine RenderEngineInterface
	logger           Logger
	httpTransport    http.RoundTripper
}

func (env *Environment) Errorf(format string, args ...interface{}) {
	if env.logger != nil {
		env.logger.Errorf(format, args)
	}
}

func (env *Environment) RenderEngineForRequest(request *http.Request) RenderEngineInterface {
	if env.mockRenderEngine != nil {
		return env.mockRenderEngine
	}
	return NewRenderEngine(env.blobStore, env.httpTransport)
}

// Use this instead of &Environment{...} directly to get compile-timer
// errors when new dependencies are introduced.
func NewEnvironment(blobStore BlobStore,
	taskQueue UrlTaskQueue,
	logger Logger,
	httpTransport http.RoundTripper) *Environment {

	return &Environment{
		blobStore:     blobStore,
		taskQueue:     taskQueue,
		logger:        logger,
		httpTransport: httpTransport,
	}
}

// It's assumed that an Environment will be request-specific (it is for Appengine),
// an EnvironmentFactory will create a new Environment for any given request.
type EnvironmentFactory interface {
	ForRequest(request *http.Request) *Environment
}

// ======================================
// ========= SYSTEM/SERVICE APIS ========
// ======================================

type Logger interface {
	Errorf(format string, args ...interface{})
}

// UrlTaskQueue
//
// Maintains, and executes a queue of tasks.  The tasks are represented as a
// base URL, plus query params.
//
// Although this removes a dependency on appengine-specific *code*, this is a
// somewhat appengine-specific *concept*.  Of course, there's no reason any
// server would choose to implement tasks this way, but I'm not sure it would
// be the first choice.  Anyway, for now we'll live with it.
type UrlTaskQueue interface {
	Enqueue(url string, params *url.Values) error
}

// ======================================
// ======= DEFAULT IMPLEMENTATIONS ======
// ======================================

// SyncUrlTaskQueue
//
// This hasn't been implemented yet, but the idea is that it would just call
// the URL over HTTP direcly, and block waiting for a response.
type SyncUrlTaskQueue struct {
	baseUrl string
}

func (q *SyncUrlTaskQueue) Enqueue(url string, params *url.Values) error {
	//	u := url.Parse(baseUrl + url + params.Encode())

	//	var req http.Request
	//	req.Method = "GET"
	//	req.header = http.Header{}
	//	req.URL = u
	panic("Not Implemented")
}

type DefaultLogger struct{}

func (l DefaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args)
}

type StaticEnvironmentFactory struct {
	staticEnvironment *Environment
}

func NewStaticEnvironmentFactory(staticEnvironment *Environment) EnvironmentFactory {
	return &StaticEnvironmentFactory{staticEnvironment: staticEnvironment}
}

func (cf *StaticEnvironmentFactory) ForRequest(request *http.Request) *Environment {
	return cf.staticEnvironment
}
