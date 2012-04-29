package latvis

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Blob struct {
	Data []byte

	// TODO(mrjones): metadata (e.g. Content-Type)
}

type Handle struct {
	timestamp  int64
	n1, n2, n3 int64
}

func (h *Handle) String() string {
	return fmt.Sprintf("%d-%d%d%d", h.timestamp, h.n1, h.n2, h.n3)
}

type BlobStore interface {
	// Stores a blob, identified by the Handle, to the BlobStore.
	// Storing a second blob with the same handle will overwrite the first one.
	Store(handle *Handle, blob *Blob) error

	// Fetches the blob with the given handle.
	// TODO(mrjones): distinguish true error from missing blob?
	Fetch(handle *Handle) (*Blob, error)
}

// ======================================
// ============ BLOB HELPERS ============
// ======================================

func generateNewHandle() *Handle {
	return &Handle{
	timestamp: time.Now().Unix(),
		n1:        rand.Int63(),
		n2:        rand.Int63(),
		n3:        rand.Int63(),
	}
}

func serializeHandleToParams(h *Handle, p *url.Values) {
	p.Add("hStamp", strconv.FormatInt(h.timestamp, 10))
	p.Add("h1", strconv.FormatInt(h.n1, 10))
	p.Add("h2", strconv.FormatInt(h.n2, 10))
	p.Add("h3", strconv.FormatInt(h.n3, 10))
}

func parseHandleFromParams(p *url.Values) (*Handle, error) {
	timestamp, err := strconv.ParseInt(p.Get("hStamp"), 10, 64)
	if err != nil {
		return nil, errors.New("[hStamp=" + p.Get("hStamp") + "]" + err.Error())
	}

	n1, err := strconv.ParseInt(p.Get("h1"), 10, 64)
	if err != nil {
		return nil, errors.New("[h1=" + p.Get("h1") + "]" + err.Error())
	}

	n2, err := strconv.ParseInt(p.Get("h2"), 10, 64)
	if err != nil {
		return nil, errors.New("[h2=" + p.Get("h2") + "]" + err.Error())
	}

	n3, err := strconv.ParseInt(p.Get("h3"), 10, 64)
	if err != nil {
		return nil, errors.New("[h3=" + p.Get("h3") + "]" + err.Error())
	}

	return &Handle{timestamp: timestamp, n1: n1, n2: n2, n3: n3}, nil
}

func serializeHandleToUrl(h *Handle, suffix string, page string) string {
	return fmt.Sprintf("/%s/%d-%d-%d-%d.%s", page, h.timestamp, h.n1, h.n2, h.n3, suffix)
}

func parseHandleFromUrl(fullpath string) (*Handle, error) {
	directories := strings.Split(fullpath, "/")
	if len(directories) != 3 {
		return nil, errors.New("Invalid filename [1]: " + fullpath)
	}
	if directories[0] != "" {
		return nil, errors.New("Invalid filename [2]: " + fullpath)
	}

	filename := directories[2]
	fileparts := strings.Split(filename, ".")

	if len(fileparts) != 2 {
		return nil, errors.New("Invalid filename [3]: " + fullpath)
	}

	pieces := strings.Split(fileparts[0], "-")
	if len(pieces) != 4 {
		return nil, errors.New("Invalid filename [4]: " + fullpath)
	}

	s, err := strconv.ParseInt(pieces[0], 10, 64)
	if err != nil {
		return nil, err
	}
	n1, err := strconv.ParseInt(pieces[1], 10, 64)
	if err != nil {
		return nil, err
	}
	n2, err := strconv.ParseInt(pieces[2], 10, 64)
	if err != nil {
		return nil, err
	}
	n3, err := strconv.ParseInt(pieces[3], 10, 64)
	if err != nil {
		return nil, err
	}
	return &Handle{timestamp: s, n1: n1, n2: n2, n3: n3}, nil
}

// ======================================
// ==== SIMPLE FLAT FILE BLOB STORE =====
// ======================================

type LocalFSBlobStore struct {
	location string
}

func NewLocalFSBlobStore(location string) *LocalFSBlobStore {
	return &LocalFSBlobStore{location: location}
}

func (s *LocalFSBlobStore) Store(handle *Handle, blob *Blob) error {
	filename := s.filename(handle)

	return ioutil.WriteFile(filename, blob.Data, 0600)
}

func (s *LocalFSBlobStore) Fetch(handle *Handle) (*Blob, error) {
	filename := s.filename(handle)
	data, err := ioutil.ReadFile(filename)
	blob := &Blob{Data: data}
	return blob, err
}

func (s *LocalFSBlobStore) filename(h *Handle) string {
	return fmt.Sprintf(s.location+"/%d-%d%d%d.png", h.timestamp, h.n1, h.n2, h.n3)
}
