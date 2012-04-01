package main

// Usage:
// - cd src/latvis
// - go install
// - go run localserver/localserver.go

import (
	"latvis"
)

func main() {
	config := latvis.NewConfig(
		&latvis.LocalFSBlobStoreProvider{},
		&latvis.StandardHttpClientProvider{},
		&latvis.InMemoryOauthSecretStoreProvider{},
		&latvis.SyncUrlTaskQueueProvider{})
	latvis.Setup(config)
	latvis.Serve()
}
