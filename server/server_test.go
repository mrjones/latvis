package server

import (
	"github.com/mrjones/oauth"
	"github.com/mrjones/gt"

	"http"
	"os"
	"testing"
	"url"
)

func TestObjectReady(t *testing.T) {
	blobStore := NewLocalFSBlobStore("testdir")
	err := os.Mkdir("testdir", 0755)
	gt.AssertNil(t, err)
	defer os.RemoveAll("testdir")

	cfg := &ServerConfig{blobStorage: &DumbBlobStoreProvider{Target: blobStore}}
	
	Setup(cfg)
	
	req, err := http.NewRequest("GET", "http://myhost.com/is_ready/100-1-2-3.png", nil)
	gt.AssertNil(t, err)

	res1 := NewFakeResponse()
	IsReadyHandler(res1, req);
	gt.AssertEqualM(t, http.StatusOK, res1.StatusCode, "Request should have succeeded")
	gt.AssertEqualM(t, "fail", res1.Body, "Should not have found the object")

	h := &Handle{n1:1, n2:2, n3:3, timestamp: 100}
	err = cfg.blobStorage.OpenStore(req).Store(h, &Blob{})
	gt.AssertNil(t, err)

	res2 := NewFakeResponse()
	IsReadyHandler(res2, req);
	gt.AssertEqualM(t, http.StatusOK, res2.StatusCode, "Request should have succeeded")
	gt.AssertEqualM(t, "ok", res2.Body, "Should have found the object this time.")
}

func TestAuthorization(t *testing.T) {
	cfg := &ServerConfig{
	secretStorage: &InMemoryOauthSecretStoreProvider{},
	httpClient: &StandardHttpClientProvider{},
	latitude: &FakeLatitudeConnector{},
	}
	Setup(cfg)

	authUrl := "http://myhost.com/authorize?lllat=1.0&lllng=2.0&urlat=3.0&urlng=4.0" +
		"&start=5&end=6"

	req, err := http.NewRequest("GET", authUrl, nil)
	gt.AssertNil(t, err)
	res := NewFakeResponse()

	AuthorizeHandler(res, req)

	// TODO(mrjones): check that the requested redirect (back to our server)
	// is passed in correctly.

	gt.AssertEqualM(t, http.StatusFound, res.StatusCode, "Should redirect")
	gt.AssertEqualM(t, "http://redirect.com", res.Headers.Get("Location"),
		"Should redirect to specified URL")
}

func TestAsyncTaskCreation(t *testing.T) {
	q := &FakeTaskQueue{}
	cfg := &ServerConfig{taskQueue: &FakeTaskQueueProvider{target: q}}
	Setup(cfg)

	u := "http://myhost.com/async_drawmap/?lllat=1.0&lllng=2.0&urlat=3.0&urlng=4.0" +
		"&start=5&end=6&oauth_token=tok&oauth_verifier=ver"
	req, err := http.NewRequest("GET", u, nil)
	gt.AssertNil(t, err)
	res := NewFakeResponse()

	AsyncDrawMapHandler(res, req);

	gt.AssertEqualM(t, http.StatusFound, res.StatusCode, "Should redirect")
	// TODO(mrjones): verify URL better.
//	gt.AssertEqualM(t, "/display/100-1-2-3.png", res.Headers.Get("Location"),
//		"Should redirect to specified URL")

	gt.AssertEqualM(t, "/drawmap_worker", q.lastUrl, "Should enqueue a drawmap worker")
	gt.AssertEqualM(t, "1.0000000000000000", q.lastParams.Get("lllat"), "token")
	gt.AssertEqualM(t, "2.0000000000000000", q.lastParams.Get("lllng"), "token")
	gt.AssertEqualM(t, "3.0000000000000000", q.lastParams.Get("urlat"), "token")
	gt.AssertEqualM(t, "4.0000000000000000", q.lastParams.Get("urlng"), "token")
	gt.AssertEqualM(t, "5", q.lastParams.Get("start"), "token")
	gt.AssertEqualM(t, "6", q.lastParams.Get("end"), "token")
	gt.AssertEqualM(t, "tok", q.lastParams.Get("oauth_token"), "token")
	gt.AssertEqualM(t, "ver", q.lastParams.Get("oauth_verifier"), "token")
}

// FakeTaskQueue
type FakeTaskQueueProvider struct {
	target *FakeTaskQueue
}
func (f *FakeTaskQueueProvider) GetQueue(req *http.Request) UrlTaskQueue {
	return f.target
}

type FakeTaskQueue struct {
	lastUrl string
	lastParams *url.Values
}
func (q *FakeTaskQueue) Enqueue(url string, params *url.Values) os.Error {
	q.lastUrl = url
	q.lastParams = params
	return nil;
}


// FakeLatitudeConnection
type FakeLatitudeConnector struct {}
func (f *FakeLatitudeConnector) NewConnection(r *http.Request) LatitudeConnection {
	return &FakeLatitudeConnection{}
}

type FakeLatitudeConnection struct { }
func (f *FakeLatitudeConnection) TokenRedirectUrl(callback string) (*oauth.RequestToken,string, os.Error) {
	return &oauth.RequestToken{Token:"TOKEN", Secret:"SECRET"}, "http://redirect.com", nil
}

//
// DumbBlobStoreProvider
//
type DumbBlobStoreProvider struct {
	Target BlobStore
}
func (p *DumbBlobStoreProvider) OpenStore(req *http.Request) BlobStore {
	return p.Target
}

//
// Fake Response
//
type FakeResponse struct {
	Headers http.Header
	Body string
	StatusCode int
}

func NewFakeResponse() *FakeResponse{
	return &FakeResponse{Headers: make(http.Header), StatusCode: -1}
}

func (r *FakeResponse) Header() http.Header { return r.Headers }
func (r *FakeResponse) WriteHeader(statusCode int) { r.StatusCode = statusCode } 
func (r *FakeResponse) Write(body []byte) (int, os.Error) {
	if r.StatusCode == -1 { r.StatusCode = 200 }
	r.Body = string(body)
	return len(body), nil
}

