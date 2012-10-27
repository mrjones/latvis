package latvis

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

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
