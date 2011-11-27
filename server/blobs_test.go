package server

import (
	"github.com/mrjones/gt"

	"testing"
)

func TestHandleString(t *testing.T) {
	h := &Handle{
		timestamp: 0,
		n1: 1,
		n2: 2,
		n3: 3,
	}

	gt.AssertEqualM(t, "0-123", h.String(), "Unexpected handle");
}
