//go:build !android

package mobile

// Stub implementations for non-Android mobile platforms.

func (w *window) updateAccessibility() {
}

func (w *window) initAccessibilityForWindow() {
}

func (w *window) cleanupAccessibilityForWindow() {
}
