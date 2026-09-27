//go:build go1.24

package container

import (
	"image/color"
	"runtime"
	"testing"
	"weak"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
)

// A ThemeOverride must not keep what it covers alive once nothing else uses it. The
// override scope is kept against every covered object, and it used to point back at the
// container, so every object ever placed under an override stayed in memory for good.
func TestThemeOverride_DoesNotRetainContent(t *testing.T) {
	test.NewTempApp(t)
	content := func() weak.Pointer[canvas.Rectangle] {
		r := canvas.NewRectangle(color.Black)
		NewThemeOverride(NewStack(r), test.Theme())
		return weak.Make(r)
	}()
	for range 3 {
		runtime.GC()
	}
	if content.Value() != nil {
		t.Error("the content of a discarded ThemeOverride is still alive")
	}
}
