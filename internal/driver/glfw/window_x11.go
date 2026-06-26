//go:build !wayland && (linux || freebsd || openbsd || netbsd) && !wasm && !test_web_driver

package glfw

import "C"
import (
	"fmt"

	"fyne.io/fyne/v2/driver"
)

// assert we are implementing driver.NativeWindow
var _ driver.NativeWindow = (*window)(nil)

func (w *window) RunNative(f func(any)) {
	context := driver.X11WindowContext{}
	if v := w.view(); v != nil {
		// Compiled for both, runtime check - run = X11, panic = Weston
		// TODO fix if this is GLFW v3.4 auto-detect, Wayland running but X code built
		// https://github.com/go-gl/glfw/issues/419
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered X11 / Wayland GLFW conflict", r)

				f(driver.WaylandWindowContext{})
			}
		}() // will panic and return without falling through

		context.WindowHandle = uintptr(v.GetX11Window())
	}

	f(context)
}
