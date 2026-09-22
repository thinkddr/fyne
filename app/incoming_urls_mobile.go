//go:build android || ios

package app

// Android and iOS deliver URLs through their native activity/delegate bridges.
func registerIncomingURLs() {}
