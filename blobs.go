package latvis

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"
)

// ======================================
// ========== BLOB STORAGE API ==========
// ======================================

type Blob struct {
	Data []byte
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

func GenerateHandle() *Handle {
	return &Handle{
		timestamp: time.Now().Unix(),
		n1:        rand.Int63(),
		n2:        rand.Int63(),
		n3:        rand.Int63(),
	}
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
