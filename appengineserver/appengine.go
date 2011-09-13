package appengineserver

import (
	"github.com/mrjones/latvis/server"
	"github.com/mrjones/oauth"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"

	"http"
	"os"
)

const (
	LATVIS_OUTPUT_DATATYPE = "latvis-output"
)

/// Blob Sorage ////
type AppengineBlobStoreProvider struct {
}

func (p *AppengineBlobStoreProvider) OpenStore(req *http.Request) server.BlobStore {
	return &AppengineBlobStore{request: req}
}

type AppengineBlobStore struct {
	request *http.Request
}

func (s *AppengineBlobStore) Store(handle *server.Handle, blob *server.Blob) os.Error {
	c := appengine.NewContext(s.request)
	c.Infof("Storing blob with handle: '%s'", handle.String())

	datastore.Put(c, keyFromHandle(handle), blob)
	return nil
}

func (s *AppengineBlobStore) Fetch(handle *server.Handle) (*server.Blob, os.Error) {
	c := appengine.NewContext(s.request)
	c.Infof("Looking up blob with handle: '%s'", handle.String())

	blob := new(server.Blob)
  if err := datastore.Get(c, keyFromHandle(handle), blob); err != nil {
		return nil, err
  }
	return blob, nil
}


/// URL/HTTP Fetching ////
type AppengineHttpClientProvider struct {
}

func (p *AppengineHttpClientProvider) GetClient(req *http.Request) oauth.HttpClient {
	c := appengine.NewContext(req)
	return urlfetch.Client(c)
}


func keyFromHandle(h *server.Handle) *datastore.Key {
	return datastore.NewKey(LATVIS_OUTPUT_DATATYPE, h.String(), 0, nil)
}

func init() {
	blobStoreProvider := &AppengineBlobStoreProvider{}
	httpClientProvider := &AppengineHttpClientProvider{}
  server.Setup(blobStoreProvider, httpClientProvider)
}
