//go:build !ci || !darwin

package animation

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2"
)

func tick(run *Runner) {
	time.Sleep(time.Millisecond * 5) // wait long enough that we are not at 0 time
	run.TickAnimations()
}

func TestGLDriver_StartAnimation(t *testing.T) {
	done := make(chan float32)
	run := &Runner{}
	a := &fyne.Animation{
		Duration: time.Millisecond * 100,
		Tick: func(d float32) {
			done <- d
		},
	}

	run.Start(a)
	go tick(run) // simulate a graphics draw loop
	select {
	case d := <-done:
		assert.Greater(t, d, float32(0))
	case <-time.After(100 * time.Millisecond):
		t.Error("animation was not ticked")
	}
}

func TestGLDriver_StopAnimation(t *testing.T) {
	done := make(chan float32)
	run := &Runner{}
	a := &fyne.Animation{
		Duration: time.Second * 10,
		Tick: func(d float32) {
			done <- d
		},
	}

	run.Start(a)
	go tick(run) // simulate a graphics draw loop
	select {
	case d := <-done:
		assert.Greater(t, d, float32(0))
	case <-time.After(time.Second):
		t.Error("animation was not ticked")
	}
	run.Stop(a)
	run.animationMutex.RLock()
	assert.Zero(t, len(run.animations))
	run.animationMutex.RUnlock()
}

func TestRunner_DurationChangedMidAnimation(t *testing.T) {
	progress := make(chan float32, 4)
	run := &Runner{}
	a := &fyne.Animation{
		Duration: time.Second,
		Curve:    fyne.AnimationLinear,
		Tick: func(d float32) {
			progress <- d
		},
	}

	run.Start(a)

	// Tick once against the original 1s duration.
	time.Sleep(20 * time.Millisecond)
	run.TickAnimations()
	var before float32
	select {
	case before = <-progress:
	case <-time.After(time.Second):
		t.Fatal("animation was not ticked")
	}

	// Shorten the duration. With the fix, the next tick must rescale to the
	// new 100ms total — the same elapsed time should now represent a much
	// larger fraction of progress. Without the fix it would still be divided
	// by the original 1s and stay close to `before`.
	a.Duration = 100 * time.Millisecond
	run.TickAnimations()
	var after float32
	select {
	case after = <-progress:
	case <-time.After(time.Second):
		t.Fatal("animation was not ticked after duration change")
	}
	// Either we jumped well above the original progress, or elapsed time was
	// already past the new duration so the animation completed on this tick.
	// Both prove the new duration was honored.
	assert.True(t, after == 1.0 || after > before*3,
		"progress should rescale to the new duration: before=%v after=%v", before, after)

	if after == 1.0 {
		return // already completed on the rescale (slow scheduler)
	}

	// Sleeping past the new duration should drive it to completion.
	time.Sleep(120 * time.Millisecond)
	run.TickAnimations()
	select {
	case d := <-progress:
		assert.Equal(t, float32(1.0), d, "animation should complete based on the new duration")
	case <-time.After(time.Second):
		t.Fatal("animation did not complete after duration change")
	}
}

func TestGLDriver_StopAnimationImmediatelyAndInsideTick(t *testing.T) {
	var wg sync.WaitGroup
	run := &Runner{}

	// stopping an animation immediately after start, should be effectively removed
	// from the internal animation list (first one is added directly to animation list)
	a := &fyne.Animation{
		Duration: time.Second,
		Tick:     func(f float32) {},
	}
	run.Start(a)
	go tick(run) // simulate a graphics draw loop
	run.Stop(a)

	run = &Runner{}
	wg = sync.WaitGroup{}

	// stopping animation inside tick function
	for i := 0; i < 10; i++ {
		wg.Add(1)
		var b *fyne.Animation
		b = &fyne.Animation{
			Duration: time.Second,
			Tick: func(d float32) {
				run.Stop(b)
				wg.Done()
			},
		}
		run.Start(b)
	}

	run = &Runner{}
	wg = sync.WaitGroup{}

	// Similar to first part, but in this time this animation should be added and then removed
	// from pendingAnimation slice.
	c := &fyne.Animation{
		Duration: time.Second,
		Tick:     func(f float32) {},
	}
	run.Start(c)
	tick(run) // simulate a graphics draw loop

	run.Stop(c)
	tick(run) // simulate a graphics draw loop

	wg.Wait()
	// animations stopped inside tick are really stopped in the next runner cycle
	time.Sleep(time.Second/60 + 100*time.Millisecond)
	run.animationMutex.RLock()
	assert.Zero(t, len(run.animations))
	run.animationMutex.RUnlock()
}
