//go:build !wasm && (linux || freebsd || openbsd || netbsd) && ((!x11 && !wayland) || wayland)

// The build constraint above intentionally matches the one go-gl/glfw uses for
// GetWaylandWindow (native_linbsd_wayland.go): GLFW 3.4 compiles its Wayland
// backend, and auto-detects the platform at runtime via glfw.GetPlatform(),
// whenever neither "x11" nor "wayland" is explicitly given (the default build)
// as well as when "wayland" is given explicitly. Gating this file on the
// literal "wayland" tag alone would miss the (likely more common) default
// build running on a real Wayland session - see issue #6080. Runtime dispatch
// (build.IsWayland, set by initGLFW once the actual platform is known) decides
// whether to actually use the Wayland gate; see newPresentGate and
// windowSurface below.

package glfw

/*
#cgo pkg-config: wayland-client
#include <stdlib.h>
#include <wayland-client.h>

// frame_state holds the presentable flag and the currently pending frame
// callback for one surface. It lives in C so no Go pointer is stored across the
// cgo boundary. We track cb so it can be destroyed on re-arm and on free,
// otherwise a callback left pending when a suspended window is closed (or
// re-armed) would leak its proxy and could fire frame_done into freed memory.
typedef struct { int ready; struct wl_callback *cb; } frame_state;

static void frame_done(void *data, struct wl_callback *cb, uint32_t t) {
    (void)t;
    frame_state *s = (frame_state *)data;
    s->ready = 1;                        // compositor presented us
    if (s->cb == cb) s->cb = NULL;       // it has fired; stop tracking it
    wl_callback_destroy(cb);
}
static const struct wl_callback_listener frame_listener = { frame_done };

static frame_state *frame_state_new(void) {
    frame_state *s = calloc(1, sizeof(frame_state));
    s->ready = 1;                        // first frame may proceed
    return s;
}
// frame_arm requests a frame callback and marks the surface not-ready. No
// commit here: the eglSwapBuffers that follows carries the request. Any
// still-pending callback (e.g. one armed while the surface was suspended) is
// destroyed first so it cannot fire later or leak.
static void frame_arm(struct wl_surface *surface, frame_state *s) {
    s->ready = 0;
    if (s->cb) wl_callback_destroy(s->cb);
    s->cb = wl_surface_frame(surface);
    wl_callback_add_listener(s->cb, &frame_listener, s);
}
static int  frame_ready(frame_state *s) { return s->ready; }
static void frame_state_free(frame_state *s) {
    if (s->cb) wl_callback_destroy(s->cb);
    free(s);
}
*/
import "C"

import (
	"unsafe"

	"fyne.io/fyne/v2/internal/build"
)

type frameTracker struct{ state *C.frame_state }

// newPresentGate returns the real Wayland gate only when the platform GLFW
// actually connected to at runtime is Wayland (build.IsWayland, set by
// initGLFW). This file is also compiled for the default (untagged) build,
// which auto-detects the platform and may end up on X11, so the choice can't
// be made at compile time alone.
func newPresentGate() presentGate {
	if !build.IsWayland {
		return noGate{}
	}
	return &frameTracker{state: C.frame_state_new()}
}

func (t *frameTracker) ready() bool { return C.frame_ready(t.state) != 0 }

func (t *frameTracker) arm(surface unsafe.Pointer) {
	if surface == nil {
		return
	}
	C.frame_arm((*C.struct_wl_surface)(surface), t.state)
}

func (t *frameTracker) markReady() { t.state.ready = 1 }

func (t *frameTracker) free() { C.frame_state_free(t.state) }

// windowSurface returns the window's *wl_surface as an opaque pointer, or nil
// if we are not actually running on Wayland. GetWaylandWindow panics if the
// platform GLFW connected to isn't Wayland (e.g. this default build fell back
// to X11 at runtime), so this must be checked here rather than relying on
// newPresentGate's choice of gate: arm's argument is evaluated eagerly at the
// call site regardless of which gate implementation is behind it.
func windowSurface(w *window) unsafe.Pointer {
	if !build.IsWayland || w.viewport == nil {
		return nil
	}
	return unsafe.Pointer(w.viewport.GetWaylandWindow())
}
