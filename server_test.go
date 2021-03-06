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

	cfg := NewEnvironment(blobStore, nil, nil, nil)

	res1 := execute(t, "http://myhost.com/is_ready/100-1-2-3.png", IsReadyHandler, cfg)
	gt.AssertEqualM(t, http.StatusOK, res1.StatusCode, "Request should have succeeded")
	gt.AssertEqualM(t, "fail", res1.Body, "Should not have found the object")

	h := &Handle{n1: 1, n2: 2, n3: 3, timestamp: 100}
	err := blobStore.Store(h, &Blob{})
	gt.AssertNil(t, err)

	res2 := execute(t, "http://myhost.com/is_ready/100-1-2-3.png", IsReadyHandler, cfg)
	gt.AssertEqualM(t, http.StatusOK, res2.StatusCode, "Request should have succeeded")
	gt.AssertEqualM(t, "ok", res2.Body, "Should have found the object this time.")
}

func TestObjectReadyMalformedUrl(t *testing.T) {
	dir, blobStore := setUpFakeBlobStore(t)
	defer os.RemoveAll(dir)

	cfg := &Environment{blobStore: blobStore}

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
//	cfg := &Environment{
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
	cfg := &Environment{taskQueue: q}
	s := "lllat=1.0&lllng=2.0&urlat%3d3.0&urlng=4.0&start=5&end=6"
	u := "http://myhost.com/async_drawmap/?code=vercode&state=" + url.QueryEscape(s)

	res := execute(t, u, AsyncDrawMapHandler, cfg)

	gt.AssertEqualM(t, http.StatusFound, res.StatusCode,
		"Should redirect. Body: "+res.Body)
	// TODO(mrjones): verify URL better.
	//	gt.AssertEqualM(t, "/display/100-1-2-3.png", res.Headers.Get("Location"),
	//		"Should redirect to specified URL")

	gt.AssertEqualM(t, "/drawmap_worker", q.lastUrl, "Should enqueue a drawmap worker")
	rawS := q.lastParams.Get("state")
	rawS, err := url.QueryUnescape(rawS)
	gt.AssertNil(t, err)
	parsedS, err := url.ParseQuery(rawS)

	gt.AssertEqualM(t, "1.0000000000000000", parsedS.Get("lllat"), "token")
	gt.AssertEqualM(t, "2.0000000000000000", parsedS.Get("lllng"), "token")
	gt.AssertEqualM(t, "3.0000000000000000", parsedS.Get("urlat"), "token")
	gt.AssertEqualM(t, "4.0000000000000000", parsedS.Get("urlng"), "token")
	gt.AssertEqualM(t, "5", parsedS.Get("start"), "token")
	gt.AssertEqualM(t, "6", parsedS.Get("end"), "token")
	gt.AssertEqualM(t, "vercode", q.lastParams.Get("verification_code"), "code")
	//	gt.AssertEqualM(t, "abc", q.lastParams.Get("access_token"), "token")
	//	gt.AssertEqualM(t, "def", q.lastParams.Get("refresh_token"), "token")
}

func TestAsyncWorker(t *testing.T) {
	mockEngine := &MockRenderEngine{}
	cfg := &Environment{mockRenderEngine: mockEngine}

	s := "lllat=1.0&lllng=2.0&urlat=3.0&urlng=4.0&start=5&end=6"
	u := "http://myhost.com/drawmap_worker/?state=" + url.QueryEscape(s) + "&access_token=abc&refresh_token=def&expiration_time=1234567890&hStamp=100&h1=1&h2=2&h3=3"

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

	gt.AssertEqualM(t, int64(100), mockEngine.lastHandle.timestamp, "")
	gt.AssertEqualM(t, int64(1), mockEngine.lastHandle.n1, "")
	gt.AssertEqualM(t, int64(2), mockEngine.lastHandle.n2, "")
	gt.AssertEqualM(t, int64(3), mockEngine.lastHandle.n3, "")

	gt.AssertEqualM(t, http.StatusOK, res.StatusCode, "")
}

func TestDisplayPage(t *testing.T) {
	cfg := &Environment{}

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
	env *Environment) *FakeResponse {
	UseEnvironmentFactory(NewStaticEnvironmentFactory(env))

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
	lastVerificationCode string
	lastRenderRequest    *RenderRequest
	lastHandle           *Handle
	blobStore            BlobStore
}

func (m *MockRenderEngine) GetOAuthUrl(callbackUrl, applicationState string) string {
	return "http://example.com/callback"
}

func (m *MockRenderEngine) FetchImage(handle *Handle) (*Blob, error) {
	if m.blobStore == nil {
		panic("No BlobStore configured!")
	} else {
		return m.blobStore.Fetch(handle)
	}
	return nil, nil
}

func (m *MockRenderEngine) Execute(renderReq *RenderRequest, verificationCode string, callbackUrl string, h *Handle) error {
	m.lastRenderRequest = renderReq
	m.lastHandle = h
	m.lastVerificationCode = verificationCode

	return nil
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
