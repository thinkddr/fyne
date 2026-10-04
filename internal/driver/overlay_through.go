package driver

import "fyne.io/fyne/v2"

// tapThrough is implemented by an overlay that, when a tap or a click dismisses it,
// lets that same tap reach what it was covering: a menu that closes on mouse down,
// not a modal dialog.
type tapThrough interface {
	TapThrough() bool
}

// scrollThrough is implemented by a non modal overlay that lets the mouse wheel
// scroll what it covers. It is refreshed afterwards, so it can follow what moved.
type scrollThrough interface {
	ScrollThrough() bool
}

// focusScope is implemented by an object that, while it returns a subtree, keeps
// keyboard focus traversal inside that subtree: a navigation drawer that traps Tab like
// a modal would, while the rest of the window stays visible and usable with the pointer.
type focusScope interface {
	FocusScope() fyne.CanvasObject
}

// PassesTapThrough reports whether a dismissed overlay lets its dismissing tap through.
func PassesTapThrough(o fyne.CanvasObject) bool {
	t, ok := o.(tapThrough)
	return ok && t.TapThrough()
}

// PassesScrollThrough reports whether an overlay lets the mouse wheel through.
func PassesScrollThrough(o fyne.CanvasObject) bool {
	s, ok := o.(scrollThrough)
	return ok && s.ScrollThrough()
}

// FocusScopeOf returns the subtree keyboard focus traversal is restricted to: the first
// visible object in `root` that declares a scope, or `root` itself.
func FocusScopeOf(root fyne.CanvasObject) fyne.CanvasObject {
	scope := root
	WalkVisibleObjectTree(root, func(o fyne.CanvasObject, _, _ fyne.Position, _ fyne.Size) bool {
		if s, ok := o.(focusScope); ok {
			if sub := s.FocusScope(); sub != nil {
				scope = sub
				return true
			}
		}
		return false
	}, nil)
	return scope
}
