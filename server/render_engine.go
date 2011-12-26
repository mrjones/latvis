package server

import (
	"github.com/mrjones/latvis/latitude"
	"github.com/mrjones/latvis/location"
	"github.com/mrjones/latvis/visualization"

	"fmt"
	"http"
	"os"
	"strconv"
	"time"
	"url"
)

// All the information necessary to specify a visualization.
type RenderRequest struct {
	bounds        *location.BoundingBox
	start, end    *time.Time
	oauthToken    string
	oauthVerifier string
}

// Serializes a RenderRequest to a url.Values so that it can be
// communicated to another URL endpoint.
func serializeRenderRequest(r *RenderRequest, m *url.Values) {
	if m == nil {
		panic("nil map")
	}

	m.Add("start", strconv.Itoa64(r.start.Seconds()))
	m.Add("end", strconv.Itoa64(r.end.Seconds()))

	m.Add("lllat", strconv.Ftoa64(r.bounds.LowerLeft().Lat, 'f', 16))
	m.Add("lllng", strconv.Ftoa64(r.bounds.LowerLeft().Lng, 'f', 16))
	m.Add("urlat", strconv.Ftoa64(r.bounds.UpperRight().Lat, 'f', 16))
	m.Add("urlng", strconv.Ftoa64(r.bounds.UpperRight().Lng, 'f', 16))

	m.Add("oauth_token", r.oauthToken)
	m.Add("oauth_verifier", r.oauthVerifier)
}

// De-Serializas a RenderRequest which has been encoded in a URL.
// It is expected that the encoding came from serializeRenderRequest.
func deserializeRenderRequest(params *url.Values) (*RenderRequest, os.Error) {
	// Parse all input parameters from the URL
	lowerLeft, err := extractCoordinateFromUrl(params, "lllat", "lllng")
	if err != nil {
		return nil, err
	}

	upperRight, err := extractCoordinateFromUrl(params, "urlat", "urlng")
	if err != nil {
		return nil, err
	}

	fmt.Printf("Bounding Box: LL[%f,%f], UR[%f,%f]",
		lowerLeft.Lat, lowerLeft.Lng, upperRight.Lat, upperRight.Lng)

	start, err := extractTimeFromUrl(params, "start")
	if err != nil {
		return nil, err
	}

	end, err := extractTimeFromUrl(params, "end")
	if err != nil {
		return nil, err
	}

	bounds, err := location.NewBoundingBox(*lowerLeft, *upperRight)
	if err != nil {
		return nil, err
	}

	oauthToken, err := extractStringFromUrl(params, "oauth_token")
	if err != nil {
		return nil, err
	}

	oauthVerifier, err := extractStringFromUrl(params, "oauth_verifier")
	if err != nil {
		return nil, err
	}

	return &RenderRequest{
		bounds:        bounds,
		start:         start,
		end:           end,
		oauthToken:    oauthToken,
		oauthVerifier: oauthVerifier,
	}, nil
}

// ======================================
// ============ URL PARSING =============
// ======================================

func extractCoordinateFromUrl(params *url.Values,
latparam string,
lngparam string) (*location.Coordinate, os.Error) {
	if params.Get(latparam) == "" {
		return nil, os.NewError("Missing required query paramter: " + latparam)
	}
	if params.Get(lngparam) == "" {
		return nil, os.NewError("Missing required query paramter: " + lngparam)
	}

	lat, err := strconv.Atof64(params.Get(latparam))
	if err != nil {
		return nil, err
	}
	lng, err := strconv.Atof64(params.Get(lngparam))
	if err != nil {
		return nil, err
	}

	return &location.Coordinate{Lat: lat, Lng: lng}, nil
}

func extractTimeFromUrl(params *url.Values, param string) (*time.Time, os.Error) {
	if params.Get(param) == "" {
		return nil, os.NewError("Missing query param: " + param)
	}
	startTs, err := strconv.Atoi64(params.Get(param))
	if err != nil {
		startTs = -1
	}
	return time.SecondsToUTC(startTs), nil
}

func extractStringFromUrl(params *url.Values, param string) (string, os.Error) {
	if params.Get(param) == "" {
		return "", os.NewError("Missing query param: " + param)
	}
	return params.Get(param), nil
}

func propogateParameter(base string, params *url.Values, key string) string {
	if params.Get(key) != "" {
		if len(base) > 0 {
			base = base + "&"
		}
		// TODO(mrjones): sigh use the right library
		base = base + key + "=" + url.QueryEscape(params.Get(key))
	}
	return base
}

// Capable of executing RenderRequests.
type RenderEngineInterface interface {
	Render(renderRequest *RenderRequest,
	httpRequest *http.Request,
	handle *Handle) os.Error
}

type RenderEngine struct {
	blobStorage           HttpBlobStoreProvider
	httpClientProvider    HttpClientProvider
	secretStorageProvider HttpOauthSecretStoreProvider
}

func (r *RenderEngine) Render(renderRequest *RenderRequest,
httpRequest *http.Request,
handle *Handle) os.Error {

	consumer := latitude.NewConsumer()
	consumer.HttpClient = r.httpClientProvider.GetClient(httpRequest)
	connection := latitude.NewConnectionForConsumer(consumer)

	rtoken := r.secretStorageProvider.GetStore(httpRequest).Lookup(
		renderRequest.oauthToken)
	if rtoken == nil {
		return os.NewError("No token stored for: " + renderRequest.oauthToken)
	}
	atoken, err := connection.ParseToken(rtoken, renderRequest.oauthVerifier)

	if err != nil {
		return err
	}

	var authorizedConnection location.HistorySource
	authorizedConnection = connection.Authorize(atoken)

	history, err := authorizedConnection.FetchRange(
		*renderRequest.start, *renderRequest.end)

	if err != nil {
		return err
	}

	data, err := visualization.Draw(
		history,
		renderRequest.bounds,
		&visualization.BWStyler{},
		512)
	if err != nil {
		return err
	}

	blob := &Blob{Data: *data}
	err = r.blobStorage.OpenStore(httpRequest).Store(handle, blob)
	if err != nil {
		return err
	}

	return nil
}
