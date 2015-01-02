package latvis

import (
	"fmt"
	"net/http"
	"time"
)

// ======================================
// ========= RENDER ENGINE API ==========
// ======================================

// The information necessary to specify a visualization.
type RenderRequest struct {
	// The geographic area, specified with a box of latitude/longitude
	// coordinates, to consider when rendering.
	Bounds *BoundingBox

	// The time period to consider when rendering.
	Start, End time.Time

	// TODO(mrjones): make this a better API
	VisualizationStyle string
}

// TODO(mrjones): I think I want to call this something like "LatvisController"
type RenderEngineInterface interface {
	// Returns a string representing an OAuth authorization URL.
	// You should redirect the user to that URL (either using an HTTP redirect,
	// or just by asking them to go to the URL manually). Your application will
	// get called back on 'callbackUrl' with the applicationState set in the
	// 'state=' query parameter, and with the OAuth verification code in the
	// 'verification_code=' parameter.
	//
	// Note: You can use 'oob' as a callbackUrl, in which case your application
	// will not be called back directly, but instead the user will be given an
	// OAuth verification code instead, which they can type into your application
	// (when using "oob", the applicationState parameter is ignored.
	GetOAuthUrl(callbackUrl string, applicationState string) string

	// Download and visualize a Latitude history.  The resulting visualization
	// will be stored using the given handle, and can be retrieved using
	// FecthImage with the same handle. Blocks until rendering is complete.
	Execute(renderRequest *RenderRequest,
		oauthVerificationCode string,
		callbackUrl string, // TODO(mrjones): make better?
		handle *Handle) error

	// Retrieve a visualization generated by 'Execute'.
	// This will return an error if the image is not ready yet.
	// TOOD(mrjones): distinguish between real error, and not-ready?
	FetchImage(handle *Handle) (*Blob, error)
}

func NewRenderEngine(blobStore BlobStore, httpTransport http.RoundTripper) RenderEngineInterface {
	return &RenderEngine{blobStore: blobStore, httpTransport: httpTransport}
}

// ======================================
// =========== IMPLEMENTATION ===========
// ======================================

type RenderEngine struct {
	blobStore     BlobStore
	httpTransport http.RoundTripper
}

func (r *RenderEngine) GetOAuthUrl(callbackUrl string, applicationState string) string {
	return GetAuthorizer(callbackUrl, r.httpTransport).StartAuthorize(callbackUrl, applicationState)
}

func (r *RenderEngine) FetchImage(handle *Handle) (*Blob, error) {
	return r.blobStore.Fetch(handle)
}

func (r *RenderEngine) Execute(renderRequest *RenderRequest,
	verificationCode string,
	callbackUrl string, // TODO(mrjones): make better?
	handle *Handle) error {

	dataStream, err := GetAuthorizer(callbackUrl, r.httpTransport).FinishAuthorize(verificationCode)
	if err != nil {
		return fmt.Errorf("FinishAuthorize failed: %s", err)
	}

	history, err := dataStream.FetchRange(renderRequest.Start, renderRequest.End)
	if err != nil {
		return fmt.Errorf("FetchRange failed: %s", err)
	}

	blob, err := r.MakeVisualization(history, renderRequest.Bounds, renderRequest.VisualizationStyle)
	if err != nil {
		return fmt.Errorf("MakeVisualization failed: %s", err)
	}

	err = r.blobStore.Store(handle, blob)
	if err != nil {
		return fmt.Errorf("Store failed: %s", err)
	}

	return nil
}

const (
	IMAGE_SIZE_PX = 512
)

func (r *RenderEngine) MakeVisualization(
	history *History, bounds *BoundingBox, style string) (*Blob, error) {
	w, h := imgSize(bounds, IMAGE_SIZE_PX)

	var visualizer Visualizer
	if (style == "svg") {
		visualizer = &SvgVisualizer{}
	} else {
		visualizer = &BwPngVisualizer{}
	}

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
