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

	t.Run("animation changes from faded to opaque", func(t *testing.T) {
		a.start()
		a.anim.Tick(0.0)
		assert.Equal(t, alpha(cursorDim), alpha(a.cursor.FillColor))
		a.anim.Tick(1.0)
		assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))
	})

	t.Run("interrupted animation is always opaque", func(t *testing.T) {
		a.interrupt()
		a.anim.Tick(0.0)
		assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))
		a.anim.Tick(0.5)
		assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))
		a.anim.Tick(1.0)
		assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))
	})

	t.Run("animation starts fading out again 300ms after interruption", func(t *testing.T) {
		timeNow = func() time.Time {
			return time.Now().Add(300 * time.Millisecond)
		}
		a.anim.Tick(0.0)
		runtime.Gosched()
		time.Sleep(10 * time.Millisecond) // ensure go routine for restart animation is executed
		a.anim.Tick(0.0)
		assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))
		a.anim.Tick(1.0)
		assert.Equal(t, alpha(cursorDim), alpha(a.cursor.FillColor))
	})

	t.Run("subsequent interruption works", func(t *testing.T) {
		timeNow = time.Now
		a.interrupt()
		a.anim.Tick(0.0)
		assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))

		timeNow = func() time.Time {
			return time.Now().Add(300 * time.Millisecond)
		}
		a.anim.Tick(0.0)
		runtime.Gosched()
		time.Sleep(10 * time.Millisecond) // ensure go routine for restart animation is executed
		a.anim.Tick(0.0)
		assert.Equal(t, alpha(cursorOpaque), alpha(a.cursor.FillColor))
		a.anim.Tick(1.0)
		assert.Equal(t, alpha(cursorDim), alpha(a.cursor.FillColor))
	})

	a.stop()
	assert.Nil(t, a.anim)
}
