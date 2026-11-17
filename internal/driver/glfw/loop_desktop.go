//go:build !wasm && !test_web_driver

package glfw

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/build"

	"github.com/go-gl/glfw/v3.4/glfw"
)

// platform values returned by forcePlatform to override GLFW's auto-detection.
const (
	platformAuto    = ""
	platformX11     = "x11"
	platformWayland = "wayland"
)

func (d *gLDriver) initGLFW() {
	switch forcePlatform() {
	case platformX11:
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformX11))
	case platformWayland:
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformWayland))
	}

	err := glfw.Init()
	if err != nil {
		fyne.LogError("failed to initialise GLFW", err)
		return
	}

	initCursors()
	if glfw.GetPlatform() == glfw.PlatformWayland {
		build.IsWayland = true
	}
}

func (d *gLDriver) pollEvents() {
	glfw.PollEvents() // This call blocks while window is being resized, which prevents freeDirtyTextures from being called
}

// waitEvents sleeps until a window event arrives, postEmptyEvent is called or the
// timeout passes, and processes the events that woke it.
func (d *gLDriver) waitEvents(timeout time.Duration) {
	glfw.WaitEventsTimeout(timeout.Seconds())
}

// postEmptyEvent wakes a waitEvents call from any goroutine.
func postEmptyEvent() {
	glfw.PostEmptyEvent()
}

// frameInterval is the shortest time between two painted frames: one refresh of the
// primary monitor (a 120 Hz screen gets 8 ms frames while something moves), never
// slower than 60 Hz nor faster than 240 Hz.
func (d *gLDriver) frameInterval() time.Duration {
	monitor := glfw.GetPrimaryMonitor()
	if monitor == nil {
		return defaultFrameInterval
	}
	mode := monitor.GetVideoMode()
	if mode == nil || mode.RefreshRate <= 0 {
		return defaultFrameInterval
	}
	interval := time.Second / time.Duration(mode.RefreshRate)
	return min(max(interval, minFrameInterval), defaultFrameInterval)
}

func (d *gLDriver) Terminate() {
	glfw.Terminate()
}
