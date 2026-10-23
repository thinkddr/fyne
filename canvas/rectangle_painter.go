// SPDX-License-Identifier: BSD-3-Clause

package canvas

import (
	"errors"
	"image"
	"sync"

	"fyne.io/fyne/v2"
)

// ErrRectanglePainterUnsupportedGL reports an unsupported supplied painter.
var ErrRectanglePainterUnsupportedGL = errors.New("canvas.Rectangle supplied painting is not supported by the GL painter")

// ErrRectanglePainterInvalidGeometry reports non-finite or negative geometry.
var ErrRectanglePainterInvalidGeometry = errors.New("canvas.Rectangle supplied painting has invalid geometry")

var rectanglePaintErrorLock sync.RWMutex

// RectanglePainter replaces software rectangle painting with an immutable plan.
// The provider owns its resources; Fyne does not close native engines or handles.
type RectanglePainter interface {
	// PaintSoftware receives the accumulated canvas pixels with their original
	// stride and coordinates. Respect destination bounds and context.Clip; do not
	// retain the destination or change it after returning. Validate unsupported
	// requests before writing: Fyne cannot roll back partial writes. No default
	// rectangle is painted on error. The ancestor clip allows outline outsets.
	PaintSoftware(destination *image.NRGBA, context RectanglePaintContext) error
}

// RectanglePaintContext describes one software paint in canvas coordinates.
type RectanglePaintContext struct {
	// Origin is the absolute logical origin, without pixel rounding.
	Origin fyne.Position
	// Size is the logical rectangle size, before scale.
	Size fyne.Size
	// Scale is the actual canvas scale.
	Scale float32
	// Clip is the ancestor and destination clip in physical pixels.
	Clip image.Rectangle
}

// PaintError returns the last supplied-painter error. Measurement and refresh
// leave it visible until a successful paint or an explicit SetPaintError(nil).
func (r *Rectangle) PaintError() error {
	rectanglePaintErrorLock.RLock()
	defer rectanglePaintErrorLock.RUnlock()
	return r.paintError
}

// SetPaintError records a supplied-painter error; nil clears it.
func (r *Rectangle) SetPaintError(err error) {
	rectanglePaintErrorLock.Lock()
	r.paintError = err
	rectanglePaintErrorLock.Unlock()
}
