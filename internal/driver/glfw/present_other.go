//go:build wasm || (!linux && !freebsd && !openbsd && !netbsd) || (x11 && !wayland)

// The complement of present_wayland.go's constraint: wasm, non-POSIX platforms
// (Windows, macOS), and explicit X11-only builds (which don't compile GLFW's
// Wayland backend at all, so there is no wl_surface to gate on). Everywhere
// else (the default build and explicit "wayland" builds) present_wayland.go
// provides the real implementation, chosen or bypassed at runtime via
// build.IsWayland.

package glfw

import "unsafe"

// newPresentGate returns the no-op gate: off Wayland entirely (X11-only,
// Windows, macOS, wasm), the render loop is unchanged.
func newPresentGate() presentGate { return noGate{} }

// windowSurface has no meaning off Wayland.
func windowSurface(_ *window) unsafe.Pointer { return nil }
