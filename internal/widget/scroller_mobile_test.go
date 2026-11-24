//go:build ci || no_glfw || android || ios || mobile

package widget

import (
	"image/color"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/internal/cache"
	"fyne.io/fyne/v2/test"
)

// withMobileScrollBars runs the test as on iOS and Android, where the bars only show
// while scrolling, with a short end delay.
func withMobileScrollBars(t *testing.T) {
	t.Helper()
	mobile, delay := mobileScrollBars, scrollEndDelay
	t.Cleanup(func() { mobileScrollBars, scrollEndDelay = mobile, delay })
	mobileScrollBars, scrollEndDelay = true, 50*time.Millisecond
	if scrollBarAlwaysVisible() && runtime.GOOS == "darwin" {
		t.Skip("the macOS preference shows the bars always")
	}
}

func TestScroll_MobileBarsShowOnlyWhileDragging(t *testing.T) {
	test.NewTempApp(t)
	withMobileScrollBars(t)

	rect := canvas.NewRectangle(color.Black)
	rect.SetMinSize(fyne.NewSize(100, 1000))
	scroll := NewVScroll(rect)
	scroll.Resize(fyne.NewSize(100, 100))
	area := cache.Renderer(scroll).(*scrollContainerRenderer).vertArea
	bar := cache.Renderer(area).(*scrollBarAreaRenderer).bar
	require.True(t, area.Visible(), "the content is taller than the scroller")
	assert.False(t, bar.Visible(), "no bar at rest")

	scroll.Dragged(&fyne.DragEvent{Dragged: fyne.Delta{DY: -10}})
	assert.Equal(t, float32(10), scroll.Offset.Y)
	assert.True(t, bar.Visible(), "the bar shows while dragging")

	// A finger resting on the list keeps the bar, however long it rests.
	scroll.Dragged(&fyne.DragEvent{})
	waitScrollEnd()
	assert.True(t, bar.Visible(), "the bar stays until the drag ends")

	scroll.DragEnd()
	assert.True(t, bar.Visible(), "the bar fades only after the delay")
	waitScrollEnd()
	assert.False(t, bar.Visible(), "the bar hides once scrolling stops")
	assert.Equal(t, float32(10), scroll.Offset.Y, "hiding the bar does not move the content")

	// A drag that starts again before the delay ends keeps the bar.
	scroll.Dragged(&fyne.DragEvent{Dragged: fyne.Delta{DY: -10}})
	scroll.DragEnd()
	scroll.Dragged(&fyne.DragEvent{Dragged: fyne.Delta{DY: -10}})
	waitScrollEnd()
	assert.True(t, bar.Visible(), "a new drag cancels the pending hide")
	scroll.DragEnd()
	waitScrollEnd()
	assert.False(t, bar.Visible())
}

// waitScrollEnd waits past scrollEndDelay and runs, on the test goroutine, what the
// end timer queued with fyne.Do.
func waitScrollEnd() {
	time.Sleep(4 * scrollEndDelay)
	fyne.DoAndWait(func() {})
}

func TestScroll_DesktopBarsStayWhileDragging(t *testing.T) {
	test.NewTempApp(t)
	delay := scrollEndDelay
	t.Cleanup(func() { scrollEndDelay = delay })
	scrollEndDelay = 50 * time.Millisecond
	if !scrollBarAlwaysVisible() {
		t.Skip("the bars of this platform only show while scrolling")
	}

	rect := canvas.NewRectangle(color.Black)
	rect.SetMinSize(fyne.NewSize(100, 1000))
	scroll := NewVScroll(rect)
	scroll.Resize(fyne.NewSize(100, 100))
	area := cache.Renderer(scroll).(*scrollContainerRenderer).vertArea
	bar := cache.Renderer(area).(*scrollBarAreaRenderer).bar
	assert.True(t, bar.Visible())

	scroll.Dragged(&fyne.DragEvent{Dragged: fyne.Delta{DY: -10}})
	scroll.DragEnd()
	waitScrollEnd()
	assert.True(t, bar.Visible())
	assert.False(t, scroll.scrolling, "no scrolling state where the bars always show")
}
