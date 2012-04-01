package appengineserver

// Configures a LatVis server which uses appengine services (e.g. blob storage,
// http client, etc.).
//
// All AppEngine-specific code should be completely encapsulated inside this package.
//
// Run it locally with:
// $ dev_appserver.py .
// From the root latvis directory.
//
// Also works in a deployed appengine instance.
import (
	"github.com/mrjones/oauth"

	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
	"appengine/urlfetch"

	"latvis"

	"http"
	"os"
	"url"
)

const (
	LATVIS_OUTPUT_DATATYPE = "latvis-output"
)

func init() {
	config := latvis.NewConfig(
		&AppengineBlobStoreProvider{},
		&AppengineHttpClientProvider{},
		&latvis.InMemoryOauthSecretStoreProvider{},
		&AppengineUrlTaskQueueProvider{})
	latvis.Setup(config)
}

//
// TASK QUEUE
//
type AppengineUrlTaskQueueProvider struct { }

func (p *AppengineUrlTaskQueueProvider) GetQueue(req *http.Request) latvis.UrlTaskQueue {
	return NewAppengineUrlTaskQueue(req)
}

type AppengineUrlTaskQueue struct {
	request *http.Request
}

func NewAppengineUrlTaskQueue(request *http.Request) latvis.UrlTaskQueue {
	return &AppengineUrlTaskQueue{request: request}
}

func (q *AppengineUrlTaskQueue) Enqueue(url string, params *url.Values) os.Error {
	c := appengine.NewContext(q.request)

	t := taskqueue.NewPOSTTask(url, *params)
	_, err := taskqueue.Add(c, t, "")
	return err
}

//
// BLOB STORAGE
//
type AppengineBlobStoreProvider struct { }

func (p *AppengineBlobStoreProvider) OpenStore(req *http.Request) latvis.BlobStore {
	return &AppengineBlobStore{request: req}
}

type AppengineBlobStore struct {
	request *http.Request
}

func (s *AppengineBlobStore) Store(handle *latvis.Handle, blob *latvis.Blob) os.Error {
	c := appengine.NewContext(s.request)
	c.Infof("Storing blob with handle: '%s'", handle.String())

	datastore.Put(c, keyFromHandle(c, handle), blob)
	return nil
}

func (s *AppengineBlobStore) Fetch(handle *latvis.Handle) (*latvis.Blob, os.Error) {
	c := appengine.NewContext(s.request)
	c.Infof("Looking up blob with handle: '%s'", handle.String())

	blob := new(latvis.Blob)
	if err := datastore.Get(c, keyFromHandle(c, handle), blob); err != nil {
		return nil, err
	}
	return blob, nil
}

//
// HTTP CLIENT
//
type AppengineHttpClientProvider struct { }

func (p *AppengineHttpClientProvider) GetClient(req *http.Request) oauth.HttpClient {
	c := appengine.NewContext(req)
	return urlfetch.Client(c)
}

func keyFromHandle(c appengine.Context, h *latvis.Handle) *datastore.Key {
	return datastore.NewKey(c, LATVIS_OUTPUT_DATATYPE, h.String(), 0, nil)
}
