//go:build !windows || !ci

package mobile

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	fynecanvas "fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"github.com/stretchr/testify/assert"
)

// dismissOnTap is an overlay that closes when tapped, and may let that tap through.
type dismissOnTap struct {
	widget.BaseWidget
	c       fyne.Canvas
	through bool
}

func (d *dismissOnTap) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(fynecanvas.NewRectangle(color.Transparent))
}
func (d *dismissOnTap) Tapped(*fyne.PointEvent) { d.c.Overlays().Remove(d) }
func (d *dismissOnTap) TapThrough() bool        { return d.through }

type tapCounter struct {
	widget.BaseWidget
	taps int
}

func (t *tapCounter) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(fynecanvas.NewRectangle(color.White))
}
func (t *tapCounter) Tapped(*fyne.PointEvent) { t.taps++ }

func Test_canvas_TapThroughDismissedOverlay(t *testing.T) {
	for _, through := range []bool{true, false} {
		content := &tapCounter{}
		content.ExtendBaseWidget(content)
		c := newCanvas(fyne.CurrentDevice()).(*canvas)
		c.SetContent(content)
		c.Resize(fyne.NewSize(100, 100))
		over := &dismissOnTap{c: c, through: through}
		over.ExtendBaseWidget(over)
		c.Overlays().Add(over)
		over.Resize(c.Size())

		pos := fyne.NewPos(50, 50)
		c.tapDown(pos, 0)
		c.tapUp(pos, 0, func(wid fyne.Tappable, ev *fyne.PointEvent) {
			wid.Tapped(ev)
		}, func(wid fyne.SecondaryTappable, ev *fyne.PointEvent) {
		}, func(wid fyne.DoubleTappable, ev *fyne.PointEvent) {
		}, func(wid fyne.Draggable, ev *fyne.DragEvent) {
		})
		assert.Nil(t, c.Overlays().Top(), "the tap dismissed the overlay")
		if through {
			assert.Equal(t, 1, content.taps, "a tap that dismisses a menu reaches what it covered")
		} else {
			assert.Equal(t, 0, content.taps, "a modal overlay keeps the tap that dismisses it")
		}
	}
}
