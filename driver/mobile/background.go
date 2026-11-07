// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package mobile

import (
	"context"
	"time"
)

// BackgroundDriver is an optional API of the mobile driver for work that the system
// starts without a user interface. Assert it on fyne.App.Driver().
//
// iOS can launch an app with no scene: for an app refresh task (BGAppRefreshTask) or
// a background notification. The whole main function runs then too, before UIKit says
// why it launched, so an app that wants to avoid user-facing work in that case defers
// it with OnFirstScene. Other platforms have no such launch: OnFirstScene runs at once,
// LaunchedInBackground is false, protected data is available and no refresh is ever
// requested.
type BackgroundDriver interface {
	// LaunchedInBackground reports whether the system launched the process without a
	// user interface and no scene has connected since. It is only meaningful once the
	// application has finished launching; before that it is false.
	LaunchedInBackground() bool

	// OnFirstScene runs f when the first scene connects: synchronously, on the main
	// thread, before the URLs or user activities of that scene are delivered, in the
	// order the functions were registered. If a scene has already connected, or the
	// platform has no scenes, f runs synchronously before OnFirstScene returns.
	// f must not block or wait for the UI goroutine.
	OnFirstScene(f func())

	// ProtectedDataAvailable reports whether files with complete data protection can
	// be read: on iOS, false while the device is locked
	// (UIApplication.protectedDataAvailable). It is safe from any goroutine.
	ProtectedDataAvailable() bool

	// OnBackgroundRefresh sets the handler of the app refresh task. The first entry of
	// BGTaskSchedulerPermittedIdentifiers in Info.plist names the task, and
	// UIBackgroundModes must contain "fetch"; without them the handler never runs.
	//
	// A refresh is requested no earlier than every from now each time the app moves to
	// the background, and again when the handler returns: after every, or after retry
	// when it is positive. f runs on its own goroutine. ctx is cancelled when the
	// system ends the task, and the task is then reported as failed as soon as f
	// returns, or after a grace of half a second if it does not. f must return
	// promptly once ctx is done, and must not hold a file lock in a shared container
	// past that: iOS suspends the app when the task completes, and kills an app that
	// is suspended holding one (0xdead10cc).
	// ok is the success reported to the system. Set it before the first scene connects
	// (in main) so that a refresh launch finds it.
	OnBackgroundRefresh(every time.Duration, f func(ctx context.Context) (ok bool, retry time.Duration))
}
