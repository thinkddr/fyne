package canvas

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
)

// ErrTextLayoutUnsupportedGL reports that a supplied text layout cannot be
// painted through the current GL texture path.
var ErrTextLayoutUnsupportedGL = errors.New("canvas.Text layout painting is not supported by the GL painter")

// ErrTextLayoutInvalidMetrics reports non-finite or negative layout dimensions.
var ErrTextLayoutInvalidMetrics = errors.New("canvas.Text layout metrics are invalid")

// ErrTextLayoutUnsupportedDecorations reports decoration geometry outside the
// software rasterizer's coordinate or allocation limits.
var ErrTextLayoutUnsupportedDecorations = errors.New("canvas.Text layout decorations exceed software raster limits")

// A shared lock protects error access without making the canvas primitive non-copyable.
var textLayoutErrorLock sync.RWMutex

// TextLayout is an immutable shaped line. Its metrics and glyph data are owned
// by the provider; Fyne does not close font engines or retain native run handles.
// Replace the layout when the text, font or shaping policy changes.
type TextLayout interface {
	// Metrics returns the line-box size and its baseline from the line-box top.
	// The baseline may be outside the line box when line-height is smaller than
	// the font metrics. All returned values must be finite; size cannot be negative.
	Metrics() (size fyne.Size, baseline float32, err error)
	// PaintSoftware paints glyphs only onto the accumulated destination. Fyne
	// paints underline and strikethrough from Metrics after a successful call.
	// These use Fyne's line stroke and baseline offsets, not CSS decoration rules.
	// The destination shares its backing storage with the canvas; respect its
	// bounds, stride and context.Clip. Do not retain or modify it after returning.
	// An unsupported request must return an error before modifying any pixels.
	// Fyne cannot roll back partial writes and never draws default glyphs on error.
	PaintSoftware(destination *image.NRGBA, context TextPaintContext) error
}

// TextPaintContext describes a software line paint in canvas coordinates.
type TextPaintContext struct {
	// Origin is the absolute logical line-box origin after text alignment.
	// It retains fractional coordinates and has not been rounded by the painter.
	Origin fyne.Position
	// Scale is the actual canvas scale for this paint.
	Scale float32
	// Clip is the intersected physical-pixel clip in destination coordinates.
	Clip image.Rectangle
	// Color is the resolved text color, including its alpha.
	Color color.Color
}

// LayoutMetrics returns the supplied line-box size and baseline, or the default
// driver metrics when Layout is nil. Invalid supplied metrics never fall back.
// A successful query clears only the metric error, leaving paint errors visible.
func (t *Text) LayoutMetrics() (fyne.Size, float32, error) {
	var size fyne.Size
	var baseline float32
	var err error
	if t.Layout == nil {
		size, baseline = fyne.CurrentApp().Driver().RenderedTextSize(t.Text, t.TextSize, t.TextStyle, t.FontSource)
	} else {
		size, baseline, err = t.Layout.Metrics()
		if err != nil {
			err = fmt.Errorf("canvas.Text layout metrics: %w", err)
		} else if size.Width < 0 || size.Height < 0 || !finiteTextMetric(size.Width) || !finiteTextMetric(size.Height) || !finiteTextMetric(baseline) {
			err = ErrTextLayoutInvalidMetrics
		}
	}
	textLayoutErrorLock.Lock()
	t.layoutMetricsError = err
	textLayoutErrorLock.Unlock()
	if err != nil {
		return fyne.Size{}, 0, err
	}
	return size, baseline, nil
}

// TextLayoutError returns the current metric and paint errors, if any.
// Calling MinSize or LayoutMetrics successfully does not clear a paint error.
// After removing Layout, call LayoutMetrics to clear a prior metric error.
func (t *Text) TextLayoutError() error {
	textLayoutErrorLock.RLock()
	defer textLayoutErrorLock.RUnlock()
	if t.layoutMetricsError == nil {
		return t.layoutPaintError
	}
	if t.layoutPaintError == nil {
		return t.layoutMetricsError
	}
	return errors.Join(t.layoutMetricsError, t.layoutPaintError)
}

// SetTextLayoutError records a paint error. Passing nil clears only that error.
// Painters call this after an attempted supplied-layout paint; providers can use
// it to explicitly clear a previous paint failure after replacing a layout.
func (t *Text) SetTextLayoutError(err error) {
	textLayoutErrorLock.Lock()
	t.layoutPaintError = err
	textLayoutErrorLock.Unlock()
}

func finiteTextMetric(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}
