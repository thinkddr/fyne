// Copyright (C) 2026 Sytue.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build darwin && !ios && !mobile && !ci && !wasm && !test_web_driver

package app

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit -framework Carbon

void watchIncomingURLs(void);
*/
import "C"

import "fyne.io/fyne/v2/internal/urlhandler"

// registerIncomingURLs installs the kAEGetURL and kAEOpenDocuments handlers before
// the application has a window. macOS sends custom-scheme launches and opened
// documents as Apple events, not argv; documents arrive as file:// URLs.
func registerIncomingURLs() { C.watchIncomingURLs() }

//export incomingURL
func incomingURL(raw *C.char) {
	if raw != nil {
		urlhandler.Deliver(C.GoString(raw))
	}
}
