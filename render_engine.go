package latvis

import (
	"net/http"
	"time"
)

const (
	IMAGE_SIZE_PX = 512
)

// All the information necessary to specify a visualization.
type RenderRequest struct {
	bounds     *BoundingBox
	start, end time.Time
}

// TODO(mrjones): I think I want to call this something like "LatvisController"
// TODO(mrjones): remove the http.Request from here
type RenderEngineInterface interface {
	GetOAuthUrl(callbackUrl string, applicationState string) string

	// Download and visualize a Latitude history.  The resulting visualization
	// will be stored using the given handle, and can be retrieved using
	// FecthImage with the same handle.
	Execute(renderRequest *RenderRequest,
		verificationCode string,
		httpRequest *http.Request,
		handle *Handle) error

	// Retrieve a visualization generated by 'Execute'.
	// This will return an error if the image is not ready yet.
	// TOOD(mrjones): distinguish between real error, and not-ready?
	FetchImage(
		handle *Handle,
		httpRequest *http.Request) (*Blob, error)
}

type RenderEngine struct {
	blobStorage HttpBlobStoreProvider
	authorizer  Authorizer
}

func (r *RenderEngine) GetOAuthUrl(callbackUrl string, applicationState string) string {
	return GetAuthorizer(callbackUrl).StartAuthorize(applicationState)
}

func (r *RenderEngine) FetchImage(handle *Handle, httpRequest *http.Request) (*Blob, error) {
	return r.blobStorage.OpenStore(httpRequest).Fetch(handle)
}

func (r *RenderEngine) Execute(renderRequest *RenderRequest,
	verificationCode string,
	httpRequest *http.Request,
	handle *Handle) error {

	dataStream, err := GetAuthorizer("TODO(mrjones): remove this?").FinishAuthorize(verificationCode)
	if err != nil {
		return err
	}

	history, err := dataStream.FetchRange(
		renderRequest.start, renderRequest.end)
	if err != nil {
		return err
	}

	blob, err := r.Render2(history, renderRequest.bounds)
	if err != nil {
		return err
	}

	err = r.blobStorage.OpenStore(httpRequest).Store(handle, blob)
	if err != nil {
		return err
	}

	return nil
}

func (r *RenderEngine) Render2(history *History, bounds *BoundingBox) (*Blob, error) {
	w, h := imgSize(bounds, IMAGE_SIZE_PX)

	visualizer := &BwPngVisualizer{}

	data, err := visualizer.Visualize(
		history,
		bounds,
		w,
		h)
	if err != nil {
		return nil, err
	}

	return &Blob{Data: *data}, nil
}

func imgSize(bounds *BoundingBox, max int) (w, h int) {
	maxF := float64(max)

	w = max
	h = max

	skew := bounds.Height() / bounds.Width()
	if skew > 1.0 {
		w = int(maxF / skew)
	} else {
		h = int(maxF * skew)
	}
	return w, h
}
