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
	"github.com/mrjones/oauth"

	"net/http"
	"net/url"
)

// ServerConfig represents all the dependencies for a latvis server.
//
// Primarily designed to separate framework-specific components (like storage)
// from the main application logic.
type ServerConfig struct {
	blobStorage   HttpBlobStoreProvider
	taskQueue     HttpUrlTaskQueueProvider
//	oauthFactory  OauthFactoryInterface
	renderEngine RenderEngineInterface
}

// Use this instead of &ServerConfig{...} directly to get compile-timer
// errors when new dependencies are introduced.
func NewConfig(blobStorage HttpBlobStoreProvider,
	
	taskQueue HttpUrlTaskQueueProvider) *ServerConfig {
	return &ServerConfig{
		blobStorage:   blobStorage,
		taskQueue:     taskQueue,
//		oauthFactory:  &RealOauthFactory{},
		renderEngine: &RenderEngine{
			blobStorage:           blobStorage,
		},

	}
}

// PROVIDERS
//
// Since Appengine libraries depend on a http.Request (indirectly, through
// appengine.Context), I've introduced these "provider" classes.  You pass
// the http.Request to a provider, and get back a class that doesn't depend
// on http.Request, meaning it can have a clean interface.  This feels more
// like Java than Go, and I'm not yet sure it was the right decision, however
// it means the other interfaces can all be clean (without dumb http.Request
// params floating everywhere, and in some cases means we can even use external
// interfaces such as oauth.HttpClient.
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

type OauthConsumerProvider interface {
	NewConsumer() *oauth.Consumer
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
	baseUrl    string
	httpClient oauth.HttpClient
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

//// StandardLatitudeConnectionProvider
//type StandardLatitudeConnector struct {
//	httpClient HttpClientProvider
//}

//func (p *StandardLatitudeConnector) NewConnection(req *http.Request) LatitudeConnection {
//	consumer := NewConsumer()
//	consumer.HttpClient = p.httpClient.GetClient(req)
//	return NewConnectionForConsumer(consumer)
//}
