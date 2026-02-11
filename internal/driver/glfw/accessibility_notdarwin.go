//go:build !darwin

package glfw

// Stub implementations for non-Darwin platforms

func (w *window) updateAccessibility() {
	// No-op on non-Darwin platforms
}

func (w *window) initAccessibilityForWindow() {
	// No-op on non-Darwin platforms
}

func (w *window) cleanupAccessibilityForWindow() {
	// No-op on non-Darwin platforms
}
