package server

import (
	"github.com/mrjones/oauth"
	"github.com/mrjones/gt"

	"http"
	"os"
	"testing"
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
		"start=5&end=6"

	req, err := http.NewRequest("GET", authUrl, nil)
	gt.AssertNil(t, err)
	res := NewFakeResponse()

	AuthorizeHandler(res, req)

	gt.AssertEqualM(t, http.StatusFound, res.StatusCode, "Should redirect")
}

// FakeLatitudeConnection
type FakeLatitudeConnector struct {}
func (f *FakeLatitudeConnector) NewConnection(r *http.Request) LatitudeConnection {
	return &FakeLatitudeConnection{}
}

type FakeLatitudeConnection struct {}
func (f *FakeLatitudeConnection) TokenRedirectUrl(callback string) (*oauth.RequestToken,string, os.Error) {
	return &oauth.RequestToken{Token:"TOKEN", Secret:"SECRET"}, "REDIRECT_URL", nil
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

