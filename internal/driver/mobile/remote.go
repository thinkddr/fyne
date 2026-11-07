// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package mobile

import (
	"context"

	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/internal/driver/mobile/app"
)

var _ mobile.RemoteNotificationDriver = (*driver)(nil)

// RemoteNotificationsAvailable implements mobile.RemoteNotificationDriver.
func (*driver) RemoteNotificationsAvailable() bool { return app.RemoteNotificationsAvailable() }

// RegisterForRemoteNotifications implements mobile.RemoteNotificationDriver.
func (*driver) RegisterForRemoteNotifications(f func(token string, err error)) {
	app.RegisterForRemoteNotifications(f)
}

// OnRemoteNotification implements mobile.RemoteNotificationDriver.
func (*driver) OnRemoteNotification(f func(ctx context.Context) (newData, ok bool)) {
	app.OnRemoteNotification(f)
}
