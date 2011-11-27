package appengineserver

import (
	"github.com/mrjones/latvis/server"
	"github.com/mrjones/oauth"

	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
	"appengine/urlfetch"

	"http"
	"os"
	"url"
)

const (
	LATVIS_OUTPUT_DATATYPE = "latvis-output"
)

type AppengineUrlTaskQueueProvider struct {
}

func (p *AppengineUrlTaskQueueProvider) GetQueue(req *http.Request) server.UrlTaskQueue {
	return NewAppengineUrlTaskQueue(req)
}

type AppengineUrlTaskQueue struct {
	request *http.Request
}

func NewAppengineUrlTaskQueue(request *http.Request) server.UrlTaskQueue {
	return &AppengineUrlTaskQueue{request: request}
}

func (q *AppengineUrlTaskQueue) Enqueue(url string, params *url.Values) os.Error {
	c := appengine.NewContext(q.request)

	t := taskqueue.NewPOSTTask(url, *params)
	_, err := taskqueue.Add(c, t, "")
	return err
}


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

	datastore.Put(c, keyFromHandle(c, handle), blob)
	return nil
}

func (s *AppengineBlobStore) Fetch(handle *server.Handle) (*server.Blob, os.Error) {
	c := appengine.NewContext(s.request)
	c.Infof("Looking up blob with handle: '%s'", handle.String())

	blob := new(server.Blob)
  if err := datastore.Get(c, keyFromHandle(c, handle), blob); err != nil {
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


func keyFromHandle(c appengine.Context, h *server.Handle) *datastore.Key {
	return datastore.NewKey(c, LATVIS_OUTPUT_DATATYPE, h.String(), 0, nil)
}

func init() {
	config := server.NewConfig(
	  &AppengineBlobStoreProvider{},
  	&AppengineHttpClientProvider{},
  	&server.InMemoryOauthSecretStoreProvider{},
		&AppengineUrlTaskQueueProvider{})
  server.Setup(config)
}
