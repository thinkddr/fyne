// Copyright (C) 2026 Sytue.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

//go:build android || ios

package app

// Android and iOS deliver URLs through their native activity/delegate bridges.
func registerIncomingURLs() {}
