//go:build !no_glfw && !mobile

package glfw

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// BenchmarkRunOnMain measures the cost of calling a function
// on the main thread.
func BenchmarkRunOnMain(b *testing.B) {
	f := func() {}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runOnMain(f)
	}
}

// BenchmarkRunOnDraw measures the cost of calling a function
// on the draw thread.
func BenchmarkRunOnDraw(b *testing.B) {
	f := func() {}
	w := createWindow("Test")
	w.create()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.RunWithContext(f)
	}
}

// TestRunOnMainWakesIdleLoop: the loop sleeps until the next event, so work sent
// from a goroutine must wake it instead of waiting for the idle timeout. Each round
// leaves the loop idle long enough to be asleep before sending.
func TestRunOnMainWakesIdleLoop(t *testing.T) {
	var slowest time.Duration
	for range 10 {
		time.Sleep(50 * time.Millisecond)
		start := time.Now()
		runOnMain(func() {})
		slowest = max(slowest, time.Since(start))
	}
	if slowest >= idleWakeInterval/2 {
		t.Fatalf("runOnMain waited %v for a sleeping loop; it should be woken at once", slowest)
	}
}

// TestDoFromGoroutineWithoutWaitWakesIdleLoop is the same for fyne.Do without
// waiting, which is how network results reach the interface.
func TestDoFromGoroutineWithoutWaitWakesIdleLoop(t *testing.T) {
	var slowest time.Duration
	for range 10 {
		time.Sleep(50 * time.Millisecond)
		ran := make(chan time.Time, 1)
		start := time.Now()
		d.DoFromGoroutine(func() { ran <- time.Now() }, false)
		select {
		case at := <-ran:
			slowest = max(slowest, at.Sub(start))
		case <-time.After(5 * time.Second):
			t.Fatal("the queued function never ran")
		}
	}
	if slowest >= idleWakeInterval/2 {
		t.Fatalf("a queued function waited %v for a sleeping loop", slowest)
	}
}

func TestNextWake(t *testing.T) {
	frame := time.Second / 60
	var idle time.Duration
	runOnMain(func() {
		// Let any pending frame of earlier tests go out first.
		for i := 0; i < 10 && d.windowNeedsFrame(); i++ {
			for _, w := range d.windows {
				w.(*window).ensurePositionProcessed()
				w.(*window).shouldExpand = false
			}
			d.drawSingleFrame()
		}
		idle = d.nextWake(time.Now(), frame)
	})
	assert.Equal(t, idleWakeInterval, idle, "with nothing to do the loop sleeps until the next event")

	w := createWindow("Next wake")
	defer w.Close()
	var soon, overdue time.Duration
	runOnMain(func() {
		w.visible = true
		w.canvas.SetDirty()
		soon = d.nextWake(time.Now(), frame)
		overdue = d.nextWake(time.Now().Add(-time.Second), frame)
	})
	assert.LessOrEqual(t, soon, frame, "a dirty canvas is painted at the next frame slot")
	assert.Greater(t, soon, time.Duration(0))
	assert.LessOrEqual(t, overdue, time.Duration(0), "an overdue frame is painted without waiting")
}
