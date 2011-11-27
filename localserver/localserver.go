package main

import (
	"github.com/mrjones/latvis/server"
)

func main() {
	config := server.NewConfig(
	  &server.LocalFSBlobStoreProvider{},
  	&server.StandardHttpClientProvider{},
  	&server.InMemoryOauthSecretStoreProvider{},
		&server.SyncUrlTaskQueueProvider{})
  server.Setup(config)
  server.Serve()     
}
