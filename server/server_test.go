package server

import (
	"github.com/mrjones/gt"

	"http"
	"os"
	"testing"
)

type FakeResponse struct {
	Headers http.Header
	Body string
	StatusCode int
}

type StubBlobStoreProvider struct {
	Target BlobStore
}
func (p *StubBlobStoreProvider) OpenStore(req *http.Request) BlobStore {
	return p.Target
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

func TestObjectReady(t *testing.T) {
	blobStore := NewLocalFSBlobStore("testdir")
	err := os.Mkdir("testdir", 0755)
	gt.AssertNil(t, err)
	defer os.RemoveAll("testdir")

	cfg := &ServerConfig{blobStorage: &StubBlobStoreProvider{Target: blobStore}}
	
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

