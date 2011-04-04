package tokens

import (
	"io/ioutil"
  "github.com/mrjones/oauth"
	"json"
	"os"
)

type Storage struct {
	filename string

	tokens map[string] *oauth.AuthorizedToken
}

func NewTokenStorage(filename string) *Storage {
	return &Storage{filename: filename, tokens: make(map[string] *oauth.AuthorizedToken)}
}

func (storage *Storage) Store(key string, token *oauth.AuthorizedToken) os.Error {
	storage.tokens[key] = token
	return storage.flush()
} 

func (storage *Storage) Fetch(key string) (*oauth.AuthorizedToken, os.Error) {
	storage.read()
	return storage.tokens[key], nil
}

func (storage *Storage) flush() os.Error {
	bytes, err := json.Marshal(storage.tokens)
	if err != nil { return err }
	return ioutil.WriteFile(storage.filename, bytes, 0666)
}

func (storage *Storage) read() os.Error {
	bytes, err := ioutil.ReadFile(storage.filename)
	if err != nil { return err }
	return json.Unmarshal(bytes, &storage.tokens)
}
