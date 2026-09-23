package test

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2"
	fynecanvas "fyne.io/fyne/v2/canvas"
)

func Test_driver_AbsolutePositionForObject(t *testing.T) {
	d := &driver{}
	w := d.CreateWindow("Test Window")
	o := fynecanvas.NewRectangle(color.Black)
	w.SetContent(o)
	w.Resize(fyne.NewSize(320, 200))

	t.Run("for padded window", func(t *testing.T) {
		w.SetPadded(true)
		assert.Equal(t, fyne.NewPos(4, 4), d.AbsolutePositionForObject(o), "safe area offset (2,3) is subtracted")
	})

	t.Run("for non-padded window", func(t *testing.T) {
		w.SetPadded(false)
		assert.Equal(t, fyne.NewPos(0, 0), d.AbsolutePositionForObject(o), "safe area offset (2,3) is subtracted")
	})
}

func TestDriver_CreateWindow(t *testing.T) {
	d := &driver{}
	w := d.CreateWindow("Test Window")

	assert.Equal(t, "Test Window", w.Title())
}

func TestDriver_DoFromGoroutine(t *testing.T) {
	NewTempApp(t)
	fromGoroutine := func(f func()) {
		sent := make(chan struct{})
		go func() {
			fyne.Do(f)
			close(sent)
		}()
		<-sent
	}

	ran := false
	fromGoroutine(func() { ran = true })
	assert.False(t, ran, "a call from another goroutine waits for the test goroutine")
	fyne.DoAndWait(func() {})
	assert.True(t, ran, "fyne.DoAndWait on the test goroutine runs what was queued first")

	ran = false
	fromGoroutine(func() { ran = true })
	NewCanvas().Capture()
	assert.True(t, ran, "a Capture runs what was queued first")

	waited := make(chan struct{})
	go func() {
		fyne.DoAndWait(func() { ran = false })
		close(waited)
	}()
	for done := false; !done; {
		Tap(&tappable{}) // an input helper runs the queue too
		select {
		case <-waited:
			done = true
		default:
		}
	}
	assert.False(t, ran, "fyne.DoAndWait from a goroutine returns once the test ran it")

	ran = false
	fromGoroutine(func() { ran = true })
	NewTempApp(t)
	fyne.DoAndWait(func() {})
	assert.False(t, ran, "a new app drops what the previous one had queued")
}

type tappable struct{ fynecanvas.Rectangle }

func (*tappable) Tapped(*fyne.PointEvent) {}
