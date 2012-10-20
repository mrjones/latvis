package latvis

import (
	"github.com/mrjones/gt"

	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestObjectReady(t *testing.T) {
	dir, blobStore := setUpFakeBlobStore(t)
	defer os.RemoveAll(dir)

	cfg := &ServerConfig{blobStorage: &DumbBlobStoreProvider{Target: blobStore}}

	res1 := execute(t, "http://myhost.com/is_ready/100-1-2-3.png", IsReadyHandler, cfg)
	gt.AssertEqualM(t, http.StatusOK, res1.StatusCode, "Request should have succeeded")
	gt.AssertEqualM(t, "fail", res1.Body, "Should not have found the object")

	h := &Handle{n1: 1, n2: 2, n3: 3, timestamp: 100}
	err := cfg.blobStorage.OpenStore(nil).Store(h, &Blob{})
	gt.AssertNil(t, err)

	res2 := execute(t, "http://myhost.com/is_ready/100-1-2-3.png", IsReadyHandler, cfg)
	gt.AssertEqualM(t, http.StatusOK, res2.StatusCode, "Request should have succeeded")
	gt.AssertEqualM(t, "ok", res2.Body, "Should have found the object this time.")
}

func TestObjectReadyMalformedUrl(t *testing.T) {
	dir, blobStore := setUpFakeBlobStore(t)
	defer os.RemoveAll(dir)

	cfg := &ServerConfig{blobStorage: &DumbBlobStoreProvider{Target: blobStore}}

	// TODO(mrjones): check error messages

	// No ".png" extension
	res := execute(t, "http://myhost.com/is_ready/100-1-2-3", IsReadyHandler, cfg)
	gt.AssertEqualM(t, http.StatusInternalServerError, res.StatusCode, "Should have been an error")

	// No "is_ready" path
	res = execute(t, "http://myhost.com/100-1-2-3.png", IsReadyHandler, cfg)
	gt.AssertEqualM(t, http.StatusInternalServerError, res.StatusCode, "Should have been an error")

	// Extraneous path
	res = execute(t, "http://myhost.com/random/is_ready/100-1-2-3.png", IsReadyHandler, cfg)
	gt.AssertEqualM(t, http.StatusInternalServerError, res.StatusCode, "Should have been an error")

	// Not enough parts in the handle
	res = execute(t, "http://myhost.com/is_ready/100-1-2.png", IsReadyHandler, cfg)
	gt.AssertEqualM(t, http.StatusInternalServerError, res.StatusCode, "Should have been an error")

	// Non-numeric parts in the handle
	res = execute(t, "http://myhost.com/is_ready/a-1-2-3.png", IsReadyHandler, cfg)
	gt.AssertEqualM(t, http.StatusInternalServerError, res.StatusCode, "Should have been an error")
}

//func TestAuthorization(t *testing.T) {
//	cfg := &ServerConfig{
//		secretStorage: &InMemoryOauthSecretStoreProvider{},
//		httpClient:    &StandardHttpClientProvider{},
//	}
//
//	authUrl := "http://myhost.com/authorize?lllat=1.0&lllng=2.0&urlat=3.0&urlng=4.0" +
//		"&start=5&end=6"
//
//	res := execute(t, authUrl, AuthorizeHandler, cfg)
//
//	// TODO(mrjones): check that the requested redirect (back to our server)
//	// is passed in correctly.
//
//	gt.AssertEqualM(t, http.StatusFound, res.StatusCode, "Should redirect")
//	gt.AssertEqualM(t, "http://redirect.com", res.Headers.Get("Location"),
//		"Should redirect to specified URL")
//}

func TestAsyncTaskCreation(t *testing.T) {
	q := &MockTaskQueue{}
	cfg := &ServerConfig{taskQueue: &MockTaskQueueProvider{target: q}}
	s := "lllat=1.0&lllng=2.0&urlat%3d3.0&urlng=4.0&start=5&end=6"
	u := "http://myhost.com/async_drawmap/?state=" + url.QueryEscape(s)

	res := execute(t, u, AsyncDrawMapHandler, cfg)

	gt.AssertEqualM(t, http.StatusFound, res.StatusCode,
		"Should redirect. Body: " + res.Body)
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

func TestAsyncWorker(t *testing.T) {
	mockEngine := &MockRenderEngine{}
	cfg := &ServerConfig{renderEngine: mockEngine}

	s := "lllat=1.0&lllng=2.0&urlat=3.0&urlng=4.0&start=5&end=6&hStamp=100&h1=1&h2=2&h3=3"
	u := "http://myhost.com/drawmap_worker/?state=" + url.QueryEscape(s)

	res := execute(t, u, DrawMapWorker, cfg)

	if mockEngine.lastRenderRequest == nil {
		t.Fatal("No render request was made!")
	}
	gt.AssertEqualM(t, 1.0, mockEngine.lastRenderRequest.bounds.LowerLeft().Lat, "")
	gt.AssertEqualM(t, 2.0, mockEngine.lastRenderRequest.bounds.LowerLeft().Lng, "")
	gt.AssertEqualM(t, 3.0, mockEngine.lastRenderRequest.bounds.UpperRight().Lat, "")
	gt.AssertEqualM(t, 4.0, mockEngine.lastRenderRequest.bounds.UpperRight().Lng, "")

	gt.AssertEqualM(t, time.Unix(5, 0).UTC(), mockEngine.lastRenderRequest.start, "")
	gt.AssertEqualM(t, time.Unix(6, 0).UTC(), mockEngine.lastRenderRequest.end, "")

	gt.AssertEqualM(t, "tok", mockEngine.lastRenderRequest.oauthToken, "")
	gt.AssertEqualM(t, "ver", mockEngine.lastRenderRequest.oauthVerifier, "")

	gt.AssertEqualM(t, int64(100), mockEngine.lastHandle.timestamp, "")
	gt.AssertEqualM(t, int64(1), mockEngine.lastHandle.n1, "")
	gt.AssertEqualM(t, int64(2), mockEngine.lastHandle.n2, "")
	gt.AssertEqualM(t, int64(3), mockEngine.lastHandle.n3, "")

	gt.AssertEqualM(t, http.StatusOK, res.StatusCode, "")
}

func TestDisplayPage(t *testing.T) {
	cfg := &ServerConfig{}

	u := "http://myhost.com/display/100-1-2-3.png"
	res := execute(t, u, ResultPageHandler, cfg)

	gt.AssertEqualM(t, http.StatusOK, res.StatusCode, "")
	gt.AssertTrueM(t, strings.Contains(res.Body, "loadImage('100-1-2-3.png'"),
		"Missing expected loadImage call in ["+res.Body+"]")
}

func setUpFakeBlobStore(t *testing.T) (string, BlobStore) {
	dir := randomDirectoryName()
	err := os.Mkdir(dir, 0755)
	gt.AssertNil(t, err)

	blobStore := NewLocalFSBlobStore(dir)
	return dir, blobStore
}

func execute(t *testing.T,
	url string,
	handler func(http.ResponseWriter, *http.Request),
	cfg *ServerConfig) *FakeResponse {
	UseConfig(cfg)

	req, err := http.NewRequest("GET", url, nil)
	gt.AssertNil(t, err)

	res := NewFakeResponse()
	handler(res, req)

	return res
}

func randomDirectoryName() string {
	return "test-dir-" + strconv.Itoa(rand.Int())
}

// MockRenderEngine
type MockRenderEngine struct {
	lastRenderRequest *RenderRequest
	lastHandle        *Handle
}

func (m *MockRenderEngine) Render(renderReq *RenderRequest, httpClient *http.Client, httpReq *http.Request, h *Handle) error {
	m.lastRenderRequest = renderReq
	m.lastHandle = h

	return nil
}

// MockTaskQueue
type MockTaskQueueProvider struct {
	target *MockTaskQueue
}

func (f *MockTaskQueueProvider) GetQueue(req *http.Request) UrlTaskQueue {
	return f.target
}

type MockTaskQueue struct {
	lastUrl    string
	lastParams *url.Values
}

func (q *MockTaskQueue) Enqueue(url string, params *url.Values) error {
	q.lastUrl = url
	q.lastParams = params
	return nil
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
	Headers    http.Header
	Body       string
	StatusCode int
}

func NewFakeResponse() *FakeResponse {
	return &FakeResponse{Headers: make(http.Header), StatusCode: -1}
}

func (r *FakeResponse) Header() http.Header        { return r.Headers }
func (r *FakeResponse) WriteHeader(statusCode int) { r.StatusCode = statusCode }
func (r *FakeResponse) Write(body []byte) (int, error) {
	if r.StatusCode == -1 {
		r.StatusCode = 200
	}
	r.Body = r.Body + string(body)
	return len(body), nil
}
