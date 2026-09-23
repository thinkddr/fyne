//go:build !no_glfw && !mobile

package glfw

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/go-gl/glfw/v3.4/glfw"
	"github.com/stretchr/testify/assert"
)

// dismissOnPress is an overlay that closes on mouse down, the way a menu does.
type dismissOnPress struct {
	widget.BaseWidget
	c            fyne.Canvas
	through      bool
	refreshCount int
}

func newDismissOnPress(c fyne.Canvas, through bool) *dismissOnPress {
	d := &dismissOnPress{c: c, through: through}
	d.ExtendBaseWidget(d)
	return d
}

func (d *dismissOnPress) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}
func (d *dismissOnPress) MouseDown(*desktop.MouseEvent) { d.c.Overlays().Remove(d) }
func (d *dismissOnPress) MouseUp(*desktop.MouseEvent)   {}
func (d *dismissOnPress) Tapped(*fyne.PointEvent)       {}
func (d *dismissOnPress) TapThrough() bool              { return d.through }
func (d *dismissOnPress) ScrollThrough() bool           { return d.through }
func (d *dismissOnPress) Refresh()                      { d.refreshCount++; d.BaseWidget.Refresh() }

func TestWindow_Tapped_RedispatchAfterTapThroughOverlayDismiss(t *testing.T) {
	for _, through := range []bool{true, false} {
		w := createWindow("Test")
		o := &tappableObject{Rectangle: canvas.NewRectangle(color.White)}
		o.SetMinSize(fyne.NewSize(100, 100))
		w.SetContent(o)
		ensureCanvasSize(t, w, fyne.NewSize(108, 108))
		over := newDismissOnPress(w.canvas, through)

		runOnMain(func() {
			w.canvas.Overlays().Add(over)
			over.Resize(w.canvas.Size())
			w.mousePos = fyne.NewPos(50, 50)
			w.mouseClicked(w.viewport, glfw.MouseButton1, glfw.Press, 0)
			w.mouseClicked(w.viewport, glfw.MouseButton1, glfw.Release, 0)

			assert.Nil(t, w.canvas.Overlays().Top(), "the press dismissed the overlay")
			if through {
				assert.NotNil(t, o.popTapEvent(), "the click that dismissed a menu reaches what it covered")
			} else {
				assert.Nil(t, o.popTapEvent(), "an overlay that does not let it through keeps its click")
			}
		})
		w.Close()
	}
}

type scrollRecorder struct {
	*canvas.Rectangle
	scrolled int
}

func (s *scrollRecorder) Scrolled(*fyne.ScrollEvent) { s.scrolled++ }

func TestWindow_Scrolled_ThroughNonModalOverlay(t *testing.T) {
	for _, through := range []bool{true, false} {
		w := createWindow("Test")
		s := &scrollRecorder{Rectangle: canvas.NewRectangle(color.White)}
		s.SetMinSize(fyne.NewSize(100, 100))
		w.SetContent(s)
		ensureCanvasSize(t, w, fyne.NewSize(108, 108))
		over := newDismissOnPress(w.canvas, through)

		runOnMain(func() {
			w.canvas.Overlays().Add(over)
			over.Resize(w.canvas.Size())
			before := over.refreshCount
			w.mousePos = fyne.NewPos(50, 50)
			w.processMouseScrolled(0, -1)
			if through {
				assert.Equal(t, 1, s.scrolled, "the wheel scrolls what a non modal overlay covers")
				assert.Greater(t, over.refreshCount, before, "and the overlay is refreshed to follow it")
			} else {
				assert.Equal(t, 0, s.scrolled, "a modal overlay keeps the wheel")
			}
		})
		w.Close()
	}
}
