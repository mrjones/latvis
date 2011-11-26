package server

import (
  "fmt"
	"io/ioutil"
	"os"
	"rand"
	"strconv"
	"strings"
	"time"
	"url"
)

type Blob struct {
	Data []byte

	// TODO(mrjones): metadata (e.g. Content-Type)
}

type Handle struct {
	timestamp int64
	n1, n2, n3 int64
}

func (h *Handle) String() string {
	return fmt.Sprintf("%d-%d%d%d", h.timestamp, h.n1, h.n2, h.n3)
}

type BlobStore interface {
	// Stores a blob, identified by the Handle, to the BlobStore.
	// Storing a second blob with the same handle will overwrite the first one.
	Store(handle *Handle, blob *Blob) os.Error

	// Fetches the blob with the given handle.
	// TODO(mrjones): distinguish true error from missing blob?
	Fetch(handle *Handle) (*Blob, os.Error)
}

// ======================================
// ============ BLOB HELPERS ============
// ======================================

func generateNewHandle() *Handle {
	return &Handle{
		timestamp: time.Seconds(),
		n1: rand.Int63(),
		n2: rand.Int63(),
		n3: rand.Int63(),
	}
}

// TODO(mrjones): generalize
func serializeHandleToUrl(h *Handle) string {
 	return fmt.Sprintf("/blob?s=%d&n1=%d&n2=%d&n3=%d", h.timestamp, h.n1, h.n2, h.n3)
}

func serializeHandleToUrl2(h *Handle, suffix string, page string) string {
 	return fmt.Sprintf("/%s/%d-%d-%d-%d.%s", page, h.timestamp, h.n1, h.n2, h.n3, suffix)
}

func serializeHandleToParams(h *Handle, p *url.Values) {
	p.Add("hStamp", strconv.Itoa64(h.timestamp))
	p.Add("h1", strconv.Itoa64(h.n1))
	p.Add("h2", strconv.Itoa64(h.n2))
	p.Add("h3", strconv.Itoa64(h.n3))
}

func parseHandleFromParams(p *url.Values) (*Handle, os.Error) {
	timestamp, err := strconv.Atoi64(p.Get("hStamp"))
	if err != nil { return nil, os.NewError("[hStamp=" + p.Get("hStamp") + "]" + err.String()) }

	n1, err := strconv.Atoi64(p.Get("h1"))
	if err != nil { return nil, os.NewError("[n1=" + p.Get("n1") + "]" + err.String()) }

	n2, err := strconv.Atoi64(p.Get("h2"))
	if err != nil { return nil, os.NewError("[n2=" + p.Get("n2") + "]" + err.String()) }

	n3, err := strconv.Atoi64(p.Get("h3"))
	if err != nil { return nil, os.NewError("[n3=" + p.Get("n3") + "]" + err.String()) }

	return &Handle{timestamp: timestamp, n1:n1, n2:n2, n3:n3}, nil
}

func parseHandle2(fullpath string) (*Handle, os.Error) {
	directories := strings.Split(fullpath, "/")
	if len(directories) != 3 {
		return nil, os.NewError("Invalid filename [1]: " + fullpath)
	}
	if directories[0] != "" {
		return nil, os.NewError("Invalid filename [2]: " + fullpath)
	}

	filename := directories[2]
	fileparts := strings.Split(filename, ".")

	if len(fileparts) != 2 {
		return nil, os.NewError("Invalid filename [3]: " + fullpath)
	}


	pieces := strings.Split(fileparts[0], "-")
	if len(pieces) != 4 {
		return nil, os.NewError("Invalid filename [4]: " + fullpath)
	}


	s, err := strconv.Atoi64(pieces[0])
	if err != nil {
		return nil, err
	}
	n1, err := strconv.Atoi64(pieces[1])
	if err != nil {
		return nil, err
	}
	n2, err := strconv.Atoi64(pieces[2])
	if err != nil {
		return nil, err
	}
	n3, err := strconv.Atoi64(pieces[3])
	if err != nil {
		return nil, err
	}
	return &Handle{timestamp: s, n1: n1, n2: n2, n3: n3}, nil
}

// ======================================
// ==== SIMPLE FLAT FILE BLOB STORE =====
// ======================================

type LocalFSBlobStore struct {
}

func (s *LocalFSBlobStore) Store(handle *Handle, blob *Blob) os.Error {
	filename := s.filename(handle)

	return ioutil.WriteFile(filename, blob.Data, 0600)
}

func (s *LocalFSBlobStore) Fetch(handle *Handle) (*Blob, os.Error) {
	filename := s.filename(handle)
	data, err := ioutil.ReadFile(filename)
	blob := &Blob{Data: data}
	return blob, err
}

func (s *LocalFSBlobStore) filename(h *Handle) string {
	return fmt.Sprintf("images/%d-%d%d%d.png", h.timestamp, h.n1, h.n2, h.n3);
}
