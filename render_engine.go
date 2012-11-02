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
	bounds        *BoundingBox
	start, end    time.Time
}

// TODO(mrjones): I think I want to call this something like "LatvisController"
type RenderEngineInterface interface {
	FetchImage(
		handle *Handle,
		httpRequest *http.Request) (*Blob, error)

	Execute(renderRequest *RenderRequest,
		dataStream DataStream,
		httpRequest *http.Request,
		handle *Handle) error
}

type RenderEngine struct {
	blobStorage           HttpBlobStoreProvider
}

func (r *RenderEngine) FetchImage(handle *Handle, httpRequest *http.Request) (*Blob, error) {
	return r.blobStorage.OpenStore(httpRequest).Fetch(handle)
}

func (r *RenderEngine) Execute(renderRequest *RenderRequest,
	dataStream DataStream,
	httpRequest *http.Request,
	handle *Handle) error {

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

	visualizer := &BwPngVisualizer{};

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
