// High level note about this file:
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
package server

import (
	"github.com/mrjones/latvis/latitude"
	"github.com/mrjones/oauth"

	"http"
	"os"
	"url"
)

// ServerConfig represents all the dependencies for a latvis server.
//
// Primarily designed to separate framework-specific components (like storage)
// from the main application logic.
type ServerConfig struct {
	blobStorage   HttpBlobStoreProvider
	httpClient    HttpClientProvider
	secretStorage HttpOauthSecretStoreProvider
	taskQueue     HttpUrlTaskQueueProvider
	latitude      HttpLatitudeConnectionProvider

	renderEngine RenderEngineInterface
}

// Use this instead of &ServerConfig{...} directly to get compile-timer
// errors when new dependencies are introduced.
func NewConfig(blobStorage HttpBlobStoreProvider,
httpClient HttpClientProvider,
secretStorage HttpOauthSecretStoreProvider,
taskQueue HttpUrlTaskQueueProvider) *ServerConfig {
	return &ServerConfig{
		blobStorage:   blobStorage,
		httpClient:    httpClient,
		secretStorage: secretStorage,
		taskQueue:     taskQueue,
		latitude: &StandardLatitudeConnector{
			httpClient: httpClient,
		},
		renderEngine: &RenderEngine{
			blobStorage:           blobStorage,
			httpClientProvider:    httpClient,
			secretStorageProvider: secretStorage,
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

type HttpClientProvider interface {
	GetClient(req *http.Request) oauth.HttpClient
}

type HttpOauthSecretStoreProvider interface {
	GetStore(req *http.Request) OauthSecretStore
}

type HttpUrlTaskQueueProvider interface {
	GetQueue(req *http.Request) UrlTaskQueue
}

type LatitudeConnection interface {
	TokenRedirectUrl(callback string) (*oauth.RequestToken, string, os.Error)
}

type HttpLatitudeConnectionProvider interface {
	NewConnection(req *http.Request) LatitudeConnection
}

// OAuthSecretStore
//
// Stores and retrieves OAuth RequestTokens.
// TODO(mrjones): there's some terminology overloading going on here that needs
//   straightening out.  In oauth.go a "RequestToken" has two parts: a "token"
//   and a "secret".  So there's a "Token" inside a "RequestToken" which is
//   confusing.  This interface would make more sense if that overloading was
//   broken.
type OauthSecretStore interface {
	Store(tokenString string, token *oauth.RequestToken)
	Lookup(tokenString string) *oauth.RequestToken
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
	Enqueue(url string, params *url.Values) os.Error
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

func (q *SyncUrlTaskQueue) Enqueue(url string, params *url.Values) os.Error {
	//	u := url.Parse(baseUrl + url + params.Encode())

	//	var req http.Request
	//	req.Method = "GET"
	//	req.header = http.Header{}
	//	req.URL = u
	panic("Not Implemented")
}

// InMemoryOauthSecretStoreProvider
//
// Stores and retrieves OAuth RequestTokens using a simple, in-memory map. This
// is fine for testing, and single-noded deployments, however it will likely
// fail if there is more than one server handling responses, since the entire
// protocol involves two calls to the latvis server.  (If one server gets the
// first call (which saves the token), and another server gets the second call
// (which looks up the token) the lookup will fail.)
type InMemoryOauthSecretStoreProvider struct {
	storage *InMemoryOauthSecretStore
}

func (p *InMemoryOauthSecretStoreProvider) GetStore(req *http.Request) OauthSecretStore {
	if p.storage == nil {
		// TODO(mrjones): lock, in case of multiple threads
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

// StandardHttpClient
//
// This implementation actually does make sense in most contexts. Appengine,
// however, is sandboxed and you can't use http.Client directly.  So for most
// non-sandboxed applications, this implementation should be sufficient.
type StandardHttpClientProvider struct{}

func (s *StandardHttpClientProvider) GetClient(req *http.Request) oauth.HttpClient {
	return nil
	//	return &http.Client{}
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

// StandardLatitudeConnectionProvider
type StandardLatitudeConnector struct {
	httpClient HttpClientProvider
}

func (p *StandardLatitudeConnector) NewConnection(req *http.Request) LatitudeConnection {
	consumer := latitude.NewConsumer()
	consumer.HttpClient = p.httpClient.GetClient(req)
	return latitude.NewConnectionForConsumer(consumer)
}
