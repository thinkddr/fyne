package desktop

// Window describes desktop specific window features
//
// Since: 2.8
type Window interface {
	// RequestFullScreenSecondary asks for a window to fullscreen on a non-primary monitor.
	// When there is only one screen it will fullscreen on the primary monitor.
	RequestFullScreenSecondary()
}
