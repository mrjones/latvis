package latvis

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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

// Serializes a RenderRequest to a url.Values so that it can be
// communicated to another URL endpoint.
func serializeRenderRequest(r *RenderRequest, m *url.Values) {
	if m == nil {
		panic("nil map")
	}
	var m2 = make(url.Values)
	
	m2.Add("start", strconv.FormatInt(r.start.Unix(), 10))
	m2.Add("end", strconv.FormatInt(r.end.Unix(), 10))

	m2.Add("lllat", strconv.FormatFloat(r.bounds.LowerLeft().Lat, 'f', 16, 64))
	m2.Add("lllng", strconv.FormatFloat(r.bounds.LowerLeft().Lng, 'f', 16, 64))
	m2.Add("urlat", strconv.FormatFloat(r.bounds.UpperRight().Lat, 'f', 16, 64))
	m2.Add("urlng", strconv.FormatFloat(r.bounds.UpperRight().Lng, 'f', 16, 64))

	m.Add("state", m2.Encode())
}

// De-Serializas a RenderRequest which has been encoded in a URL.
// It is expected that the encoding came from serializeRenderRequest.
func deserializeRenderRequest(rawParams *url.Values) (*RenderRequest, error) {
	stateString, err := url.QueryUnescape(rawParams.Get("state"))
	if err != nil {
		return nil, err
	}
	fmt.Println("state string: " + stateString)
	params, err := url.ParseQuery(stateString)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("llat? " + params.Get("lllat"))
	

	// Parse all input parameters from the URL
	lowerLeft, err := extractCoordinateFromUrl(&params, "lllat", "lllng")
	if err != nil {
		return nil, err
	}

	upperRight, err := extractCoordinateFromUrl(&params, "urlat", "urlng")
	if err != nil {
		return nil, err
	}

	fmt.Printf("Bounding Box: LL[%f,%f], UR[%f,%f]",
		lowerLeft.Lat, lowerLeft.Lng, upperRight.Lat, upperRight.Lng)

	start, err := extractTimeFromUrl(&params, "start")
	if err != nil {
		return nil, err
	}

	end, err := extractTimeFromUrl(&params, "end")
	if err != nil {
		return nil, err
	}

	bounds, err := NewBoundingBox(*lowerLeft, *upperRight)
	if err != nil {
		return nil, err
	}

	return &RenderRequest{
		bounds:        bounds,
		start:         start,
		end:           end,
	}, nil
}

// ======================================
// ============ URL PARSING =============
// ======================================

func extractCoordinateFromUrl(params *url.Values,
	latparam string,
	lngparam string) (*Coordinate, error) {
	if params.Get(latparam) == "" {
		return nil, errors.New("Missing required query paramter: " + latparam)
	}
	if params.Get(lngparam) == "" {
		return nil, errors.New("Missing required query paramter: " + lngparam)
	}

	lat, err := strconv.ParseFloat(params.Get(latparam), 64)
	if err != nil {
		return nil, err
	}
	lng, err := strconv.ParseFloat(params.Get(lngparam), 64)
	if err != nil {
		return nil, err
	}

	return &Coordinate{Lat: lat, Lng: lng}, nil
}

func extractTimeFromUrl(params *url.Values, param string) (time.Time, error) {
	if params.Get(param) == "" {
		return time.Now(), errors.New("Missing query param: " + param)
	}
	startTs, err := strconv.ParseInt(params.Get(param), 10, 64)
	if err != nil {
		startTs = -1
	}
	return time.Unix(startTs, 0).UTC(), nil
}

func extractStringFromUrl(params *url.Values, param string) (string, error) {
	if params.Get(param) == "" {
		return "", errors.New("Missing query param: " + param)
	}
	return params.Get(param), nil
}

func propogateParameter(base string, params *url.Values, key string) string {
	if params.Get(key) != "" {
		if len(base) > 0 {
			base = base + "&"
		}
		// TODO(mrjones): sigh use the right library
//		base = base + key + "=" + url.QueryEscape(params.Get(key))
		base = base + key + "=" + params.Get(key)
	}
	return base
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

	data, err := Draw(
		history,
		bounds,
		&BWStyler{},
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
