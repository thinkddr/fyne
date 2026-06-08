package widget

import (
	"image/color"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2/canvas"
	col "fyne.io/fyne/v2/internal/color"
	_ "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestEntryCursorAnim(t *testing.T) {
	cursorOpaque := theme.Color(theme.ColorNamePrimary)
	r, g, b, _ := col.ToNRGBA(cursorOpaque)
	cursorDim := color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0x16}

	alpha := func(c color.Color) uint8 {
		_, _, _, a := col.ToNRGBA(c)
		return uint8(a >> 8) // only check 8bit colour channels
	}

	cursor := canvas.NewRectangle(color.Black)
	a := newEntryCursorAnimation(cursor)

	a.start()
	a.anim.Tick(0.0)
	assert.Equal(t, alpha(cursorDim), alpha(a.cursor.FillColor))
	a.anim.Tick(1.0)
	assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))

	a.interrupt()
	a.anim.Tick(0.0)
	assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))
	a.anim.Tick(0.5)
	assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))
	a.anim.Tick(1.0)
	assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))

	timeNow = func() time.Time {
		return time.Now().Add(cursorInterruptTime)
	}
	// animation should be restarted inverting the colors
	a.anim.Tick(0.0)
	runtime.Gosched()
	time.Sleep(10 * time.Millisecond) // ensure go routine for restart animation is executed
	a.anim.Tick(0.0)
	assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))
	a.anim.Tick(1.0)
	assert.Equal(t, alpha(cursorDim), alpha(a.cursor.FillColor))

	timeNow = time.Now
	a.interrupt()
	a.anim.Tick(0.0)
	assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))

	timeNow = func() time.Time {
		return time.Now().Add(cursorInterruptTime)
	}
	a.anim.Tick(0.0)
	runtime.Gosched()
	time.Sleep(10 * time.Millisecond) // ensure go routine for restart animation is executed
	a.anim.Tick(0.0)
	assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))
	a.anim.Tick(1.0)
	assert.Equal(t, alpha(cursorDim), alpha(a.cursor.FillColor))

	a.stop()
	assert.Nil(t, a.anim)
}
