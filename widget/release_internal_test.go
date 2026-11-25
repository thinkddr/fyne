package widget

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/cache"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/test"

	"github.com/stretchr/testify/assert"
)

func vbox(objs ...fyne.CanvasObject) *fyne.Container {
	return &fyne.Container{Layout: layout.NewVBoxLayout(), Objects: objs}
}

type countDestroy struct {
	BaseWidget
	child     fyne.CanvasObject
	destroyed int
}

func newCountDestroy(child fyne.CanvasObject) *countDestroy {
	c := &countDestroy{child: child}
	c.ExtendBaseWidget(c)
	return c
}

func (c *countDestroy) CreateRenderer() fyne.WidgetRenderer {
	return &countDestroyRenderer{WidgetRenderer: NewSimpleRenderer(c.child), w: c}
}

type countDestroyRenderer struct {
	fyne.WidgetRenderer
	w *countDestroy
}

func (r *countDestroyRenderer) Destroy() {
	r.w.destroyed++
	r.WidgetRenderer.Destroy()
}

func TestReleaseRenderers_FreesTheReplacedTreeAndKeepsTheReused(t *testing.T) {
	test.NewTempApp(t)

	label := NewLabel("old")
	inner := newCountDestroy(label)
	outer := newCountDestroy(inner)
	reused := NewButton("kept", nil)
	old := vbox(outer, reused)
	w := test.NewTempWindow(t, old)
	w.Resize(fyne.NewSize(200, 200))
	w.Canvas().Capture()
	for _, wid := range []fyne.Widget{outer, inner, label, reused} {
		assert.True(t, cache.IsRendered(wid))
	}

	replacement := vbox(NewLabel("new"), reused)
	w.SetContent(replacement)
	w.Canvas().Capture()
	ReleaseRenderers(old, replacement)

	assert.False(t, cache.IsRendered(outer))
	assert.False(t, cache.IsRendered(inner), "the walk goes down through cached renderers")
	assert.False(t, cache.IsRendered(label))
	assert.Equal(t, 1, outer.destroyed)
	assert.Equal(t, 1, inner.destroyed)
	assert.True(t, cache.IsRendered(reused), "a widget that is also in a kept tree is left alone")

	// Shown again, a released widget simply gets a new renderer.
	w.SetContent(old)
	w.Canvas().Capture()
	assert.True(t, cache.IsRendered(outer))
	assert.True(t, cache.IsRendered(label))
}

func TestReleaseRenderers_CreatesNoRenderer(t *testing.T) {
	test.NewTempApp(t)

	never := newCountDestroy(NewLabel("never shown"))
	ReleaseRenderers(vbox(never))
	assert.False(t, cache.IsRendered(never))
	assert.Equal(t, 0, never.destroyed)

	ReleaseRenderers(nil)
}
