// TODO(mrjones): reconcile with server.OauthSecretStore
package latvis

import (
	"github.com/mrjones/oauth"

	"encoding/json"
	"io/ioutil"
)

type Storage struct {
	filename string

	tokens map[string]*oauth.AccessToken
}

func NewTokenStorage(filename string) *Storage {
	return &Storage{filename: filename, tokens: make(map[string]*oauth.AccessToken)}
}

func (storage *Storage) Store(key string, token *oauth.AccessToken) error {
	storage.tokens[key] = token
	return storage.flush()
}

func (storage *Storage) Fetch(key string) (*oauth.AccessToken, error) {
	storage.read()
	return storage.tokens[key], nil
}

func (storage *Storage) flush() error {
	bytes, err := json.Marshal(storage.tokens)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(storage.filename, bytes, 0666)
}

func (storage *Storage) read() error {
	bytes, err := ioutil.ReadFile(storage.filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &storage.tokens)
}
