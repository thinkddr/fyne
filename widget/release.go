package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/cache"
)

// ReleaseRenderers frees at once what the renderer cache holds for old and for
// everything it contains: the renderers of its widgets (calling their Destroy), the
// canvas each object was drawn on and their textures. Without it the cache keeps all of
// that until it expires, a minute after old was last painted, so an app that rebuilds a
// screen on every change holds a minute of old screens.
//
// It is meant for a tree that has been replaced and will not be shown again. A widget
// in it that is shown later is not broken: it gets a new renderer. Anything that is
// also in one of the kept trees, which is where an app that rebuilds only part of a
// screen leaves the pieces it reuses, is left as it is, together with everything it
// contains.
//
// Only containers and the renderers that already exist are walked: no renderer is
// created. It must be called on the goroutine that runs the app's UI.
//
// Since: 2.8
func ReleaseRenderers(old fyne.CanvasObject, kept ...fyne.CanvasObject) {
	var k cache.Kept
	if len(kept) > 0 {
		k = cache.Kept{}
		for _, o := range kept {
			k.Keep(o)
		}
	}
	cache.ReleaseTree(old, k)
}
