package async

import (
	"github.com/mrjones/latvis/location"

	"fmt"
	"time"
)

// TODO(mrjones): reconcile with server.RenderRequest
type AsyncTask struct {
	id string

	bounds *location.BoundingBox
	start, end *time.Time
	oauthToken string
	oauthVerifier string
}

func dummy() {
	fmt.Println("Hello world")
}
