//go:build !darwin

package glfw

import (
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
)

const desktopDefaultDoubleTapDelay = 300 * time.Millisecond

func (w *window) doSetFullScreen(full bool) {
	monitor := w.getMonitorForWindow()

	w.doApplyFullScreen(monitor, full)
}

func (w *window) doSetFullScreen2(full bool) {
	primary := glfw.GetPrimaryMonitor()
	monitor := primary
	for _, m := range glfw.GetMonitors() {
		if m.GetName() != primary.GetName() {
			monitor = m
			break
		}
	}

	w.doApplyFullScreen(monitor, full)
}

func (w *window) doApplyFullScreen(monitor *glfw.Monitor, full bool) {
	mode := monitor.GetVideoMode()
	if full {
		w.viewport.SetMonitor(monitor, 0, 0, mode.Width, mode.Height, mode.RefreshRate)
	} else {
		if w.width == 0 && w.height == 0 { // if we were fullscreen on creation...
			s := w.canvas.Size().Max(w.canvas.MinSize())
			w.width, w.height = w.screenSize(s)
		}
		w.viewport.SetMonitor(nil, w.xpos, w.ypos, w.width, w.height, 0)
	}
}
