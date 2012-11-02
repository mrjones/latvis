package latvis

import (
)

// Turns a location history into an image, based on the selected style.
// - history:   The list of points to render.
// - bounds:    The borders of the image (points outside the bounds are dropped).
// - styler:    The Styler to use when turning the history into an image
// - width/height:  The width & height of the final image in pixels
//
// returns
// - a []byte representing a PNG image
// TODO(mrjones): kill this
func Draw(history *History, bounds *BoundingBox, styler Styler, width, height int) (*[]byte, error) {

	return styler.Style(history, bounds, width, height)
}

