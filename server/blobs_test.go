package server

import (
	"github.com/mrjones/gt"

	"testing"
	"url"
)

func simpleHandle() *Handle {
	return &Handle{
		timestamp: 0,
		n1: 1,
		n2: 2,
		n3: 3,
	}

}

func TestHandleDebugString(t *testing.T) {
	h := simpleHandle()

	gt.AssertEqualM(t, "0-123", h.String(), "Unexpected handle");
}

func TestHandleUrlString(t *testing.T) {
	h := simpleHandle()

	gt.AssertEqualM(t, 
		"/page/0-1-2-3.xyz",
		serializeHandleToUrl2(h, "xyz", "page"),
		"Unexpected serialization")
}

func TestHandleToParams(t *testing.T) {
	h := simpleHandle()

	var p = make(url.Values)
	
	serializeHandleToParams(h, &p)
	h2, err := parseHandleFromParams(&p)

	gt.AssertNil(t, err)
	gt.AssertEqualM(t, h, h2, "Should be equal")
		
}
