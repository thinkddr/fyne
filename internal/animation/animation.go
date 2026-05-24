package animation

import (
	"time"

	"fyne.io/fyne/v2"
)

type anim struct {
	a           *fyne.Animation
	repeatsLeft int
	reverse     bool
	start       time.Time
	stopped     bool
}

func newAnim(a *fyne.Animation) *anim {
	animate := &anim{a: a, start: time.Now()}
	animate.repeatsLeft = a.RepeatCount
	return animate
}

func (a *anim) setStopped() {
	a.stopped = true
}

func (a *anim) isStopped() bool {
	return a.stopped
}
