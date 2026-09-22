// Copyright (C) 2026 Sytue.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build (!darwin || mobile || ci || wasm || test_web_driver || tinygo) && !android && !ios

package app

// registerIncomingURLs lets platform drivers install their native URL callback.
// The generic desktop path is handled from os.Args in newAppWithDriver.
func registerIncomingURLs() {}
