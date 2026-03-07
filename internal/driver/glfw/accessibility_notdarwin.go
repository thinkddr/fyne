//go:build !darwin && !windows

package glfw

func (w *window) updateAccessibility() {
}

func (w *window) initAccessibilityForWindow() {
}

func (w *window) cleanupAccessibilityForWindow() {
}
