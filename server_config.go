// High level note about this file
//
// Using appengine services (datastore, urlfetcher) need an appengine.Context
// which requires the http.Request at construction time.
//
// These interfaces are for isolating the appengine specific code, but are still
// awkward since they require an http.Request to construct seemingly unrelated
// objects.
//
// Some default implementations are also provided in this file, however, a
// number of them are only for testing and shouldn't be used in deployed
// servers.
package latvis

import (
	"log"
	"net/http"
	"net/url"
)

type EnvironmentFactory interface {
	ForRequest(request *http.Request) *Environment
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

// Environment represents all the dependencies for a latvis server.
//
// Primarily designed to separate framework-specific components (like storage)
// from the main application logic.
type Environment struct {
	blobStorage  HttpBlobStoreProvider
	taskQueue    HttpUrlTaskQueueProvider
	mockRenderEngine RenderEngineInterface
	logger       Logger
	httpTransport http.RoundTripper
}

type Logger interface {
	Errorf(format string, args ...interface{})
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
	return NewRenderEngine(env.blobStorage.OpenStore(request), env.httpTransport)
}


// Use this instead of &Environment{...} directly to get compile-timer
// errors when new dependencies are introduced.
func NewEnvironment(blobStorage HttpBlobStoreProvider,
	taskQueue HttpUrlTaskQueueProvider,
	logger Logger) *Environment {
	return &Environment{blobStorage: blobStorage, taskQueue: taskQueue, logger: logger}
}

type DefaultLogger struct { }

func (l DefaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args)
}

// PROVIDERS
//
// Since Appengine libraries depend on a http.Request (indirectly, through
// appengine.Context), I've introduced these "provider" classes.  You pass
// the http.Request to a provider, and get back a class that doesn't depend
// on http.Request, meaning it can have a clean interface.  This feels more
// like Java than Go, and I'm not yet sure it was the right decision, however
// it means the other interfaces can all be clean (without dumb http.Request
// params floating everywhere).
type HttpBlobStoreProvider interface {
	// BlobStore is defined in blobs.go
	OpenStore(req *http.Request) BlobStore
}

type HttpUrlTaskQueueProvider interface {
	GetQueue(req *http.Request) UrlTaskQueue
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

// DEFAULT IMPLEMENTATIONS
//
// Provided when running outside of the appengine framework, these are mostly
// simple implementations that can be used for testing, but might not make
// sense for a deployed, production server

// SyncUrlTaskQueue
//
// This hasn't been implemented yet, but the idea is that it would just call
// the URL over HTTP direcly, and block waiting for a response.
type SyncUrlTaskQueueProvider struct{}

func (p *SyncUrlTaskQueueProvider) GetQueue(req *http.Request) UrlTaskQueue {
	panic("You need to implement me")
}

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

// LocalFsBlobStore
//
// Defers all the work to LocalFsBlobStore in blobs.go
type LocalFSBlobStoreProvider struct {
	Location string
}

func (p *LocalFSBlobStoreProvider) OpenStore(req *http.Request) BlobStore {
	return NewLocalFSBlobStore(p.Location)
}
