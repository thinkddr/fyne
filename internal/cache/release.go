package cache

import "fyne.io/fyne/v2"

// Kept is the set of objects that ReleaseTree must leave alone.
type Kept map[fyne.CanvasObject]struct{}

// Keep adds o, and everything it contains, to k. Only containers and the cached
// renderers of widgets are walked: no renderer is created.
func (k Kept) Keep(o fyne.CanvasObject) {
	if o == nil {
		return
	}
	if _, ok := k[o]; ok {
		return
	}
	k[o] = struct{}{}
	switch x := o.(type) {
	case *fyne.Container:
		for _, c := range x.Objects {
			k.Keep(c)
		}
	case fyne.Widget:
		rinfo, ok := renderers.Load(superOf(x))
		if !ok || rinfo == nil {
			return
		}
		for _, c := range rinfo.renderer.Objects() {
			k.Keep(c)
		}
	}
}

// ReleaseTree destroys at once the cached renderers of o and of everything it contains,
// forgets their canvas and marks their textures as expired, so that the next texture
// clean on the painter's thread frees them. It does what the cache would otherwise do
// a minute after the tree stopped being painted. Objects in kept, and what they
// contain, are left as they are. kept may be nil.
func ReleaseTree(o fyne.CanvasObject, kept Kept) {
	if o == nil {
		return
	}
	if _, ok := kept[o]; ok {
		return
	}
	canvases.Delete(o)
	if tinfo, ok := objectTextures.Load(o); ok && tinfo != nil {
		tinfo.expires.Store(0)
	}
	// Every object placed under a ThemeOverride has an entry for its scope, primitives
	// and containers too, and only a widget's was deleted (with its renderer).
	overrides.Delete(o)
	switch x := o.(type) {
	case *fyne.Container:
		for _, c := range x.Objects {
			ReleaseTree(c, kept)
		}
	case fyne.Widget:
		wid := superOf(x)
		rinfo, ok := renderers.LoadAndDelete(wid)
		overrides.Delete(wid)
		if !ok || rinfo == nil {
			return
		}
		objs := rinfo.renderer.Objects()
		rinfo.renderer.Destroy()
		for _, c := range objs {
			ReleaseTree(c, kept)
		}
	}
}

// superOf is the widget that extends wid, which is the key of its renderer.
func superOf(wid fyne.Widget) fyne.Widget {
	if wd, ok := wid.(isBaseWidget); ok && wd.super() != nil {
		return wd.super()
	}
	return wid
}
