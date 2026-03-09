package desktop

// Window describes desktop specific window features
//
// Since: 2.8
type Window interface {
	// RequestFullScreenSecondary asks for a window to fullscreen on a non-primary monitor.
	// When there is only one screen it will fullscreen on the primary monitor.
	RequestFullScreenSecondary()

	// RequestAlwaysOnTop asks for the window to remain above other windows, call before [Show].
	// This is subject to support on the current system and with the caveat that the window manager
	// may decide that other windows or panels remain above this one.
	RequestAlwaysOnTop()
}
