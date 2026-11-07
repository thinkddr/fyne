// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package mobile

import "context"

// RemoteNotificationDriver is an optional API of the mobile driver for background
// remote notifications: pushes with content-available and no alert, that wake the app so
// it can fetch what changed. Assert it on fyne.App.Driver().
//
// Only iOS has them, and only for an app whose Info.plist lists "remote-notification" in
// UIBackgroundModes and that is signed with the aps-environment entitlement. Elsewhere
// they are not available: registering fails at once and the handler never runs.
type RemoteNotificationDriver interface {
	// RemoteNotificationsAvailable reports whether the app declares the
	// remote-notification background mode. It is safe from any goroutine and before
	// the application has finished launching.
	RemoteNotificationsAvailable() bool

	// RegisterForRemoteNotifications asks the system for the device token of this app,
	// and calls f once, on its own goroutine, with the token in lowercase hexadecimal or
	// with the error the system gave. It shows no permission prompt: background
	// notifications need none. Calling it again is cheap and may return the same token;
	// the system can change it, so register on every launch.
	RegisterForRemoteNotifications(f func(token string, err error))

	// OnRemoteNotification sets the handler of background remote notifications. f runs
	// on its own goroutine with the payload left out: it should fetch what it needs from
	// its own servers. ctx is cancelled after RemoteNotificationTime, and the
	// notification is then reported as failed as soon as f returns, or after a grace of
	// half a second if it does not: iOS gives the app about 30 seconds in all.
	// newData is whether the app fetched something new; ok, whether it worked. Set it
	// before the first scene connects (in main) so that a launch for a notification
	// finds it; one that arrives with no handler fails at once.
	OnRemoteNotification(f func(ctx context.Context) (newData, ok bool))
}
