package latvis

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ======================================
// =========== RENDER REQUEST ===========
// ======================================

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


// ======================================
// ============== HANDLES ===============
// ======================================

func serializeHandleToParams(h *Handle, p *url.Values) {
	p.Add("hStamp", strconv.FormatInt(h.timestamp, 10))
	p.Add("h1", strconv.FormatInt(h.n1, 10))
	p.Add("h2", strconv.FormatInt(h.n2, 10))
	p.Add("h3", strconv.FormatInt(h.n3, 10))
}

func parseHandleFromParams(p *url.Values) (*Handle, error) {
	timestamp, err := strconv.ParseInt(p.Get("hStamp"), 10, 64)
	if err != nil {
		return nil, errors.New("[hStamp=" + p.Get("hStamp") + "]" + err.Error())
	}

	n1, err := strconv.ParseInt(p.Get("h1"), 10, 64)
	if err != nil {
		return nil, errors.New("[h1=" + p.Get("h1") + "]" + err.Error())
	}

	n2, err := strconv.ParseInt(p.Get("h2"), 10, 64)
	if err != nil {
		return nil, errors.New("[h2=" + p.Get("h2") + "]" + err.Error())
	}

	n3, err := strconv.ParseInt(p.Get("h3"), 10, 64)
	if err != nil {
		return nil, errors.New("[h3=" + p.Get("h3") + "]" + err.Error())
	}

	return &Handle{timestamp: timestamp, n1: n1, n2: n2, n3: n3}, nil
}

func serializeHandleToUrl(h *Handle, suffix string, page string) string {
	return fmt.Sprintf("/%s/%d-%d-%d-%d.%s", page, h.timestamp, h.n1, h.n2, h.n3, suffix)
}

func parseHandleFromUrl(fullpath string) (*Handle, error) {
	directories := strings.Split(fullpath, "/")
	if len(directories) != 3 {
		return nil, errors.New("Invalid filename [1]: " + fullpath)
	}
	if directories[0] != "" {
		return nil, errors.New("Invalid filename [2]: " + fullpath)
	}

	filename := directories[2]
	fileparts := strings.Split(filename, ".")

	if len(fileparts) != 2 {
		return nil, errors.New("Invalid filename [3]: " + fullpath)
	}

	pieces := strings.Split(fileparts[0], "-")
	if len(pieces) != 4 {
		return nil, errors.New("Invalid filename [4]: " + fullpath)
	}

	s, err := strconv.ParseInt(pieces[0], 10, 64)
	if err != nil {
		return nil, err
	}
	n1, err := strconv.ParseInt(pieces[1], 10, 64)
	if err != nil {
		return nil, err
	}
	n2, err := strconv.ParseInt(pieces[2], 10, 64)
	if err != nil {
		return nil, err
	}
	n3, err := strconv.ParseInt(pieces[3], 10, 64)
	if err != nil {
		return nil, err
	}
	return &Handle{timestamp: s, n1: n1, n2: n2, n3: n3}, nil
}
