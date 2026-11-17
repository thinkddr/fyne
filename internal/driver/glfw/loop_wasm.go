//go:build wasm || test_web_driver

package glfw

import (
	"time"

	"fyne.io/fyne/v2"

	"github.com/fyne-io/gl-js"
	"github.com/fyne-io/glfw-js"
)

func (d *gLDriver) initGLFW() {
	err := glfw.Init(gl.ContextWatcher)
	if err != nil {
		fyne.LogError("failed to initialise GLFW", err)
		return
	}
}

func (d *gLDriver) pollEvents() {
	glfw.PollEvents() // This call blocks while window is being resized, which prevents freeDirtyTextures from being called
}

// waitEvents cannot block the browser's event loop: it sleeps (which yields to it)
// for at most one frame and then polls, as the ticker-driven loop did.
func (d *gLDriver) waitEvents(timeout time.Duration) {
	time.Sleep(min(timeout, defaultFrameInterval))
	glfw.PollEvents()
}

// postEmptyEvent has nothing to interrupt in the browser: waitEvents never sleeps
// longer than a frame.
func postEmptyEvent() {}

func (d *gLDriver) frameInterval() time.Duration {
	return defaultFrameInterval
}

func (d *gLDriver) Terminate() {
	glfw.Terminate()
}
