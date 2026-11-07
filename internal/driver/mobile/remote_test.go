// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package mobile

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/driver/mobile"
)

// Apps reach the remote notification API by asserting the driver they were given, and
// without the background mode it is not available.
func TestTheDriverOffersRemoteNotifications(t *testing.T) {
	d, ok := NewGoMobileDriver().(mobile.RemoteNotificationDriver)
	if !ok {
		t.Fatal("the mobile driver does not implement mobile.RemoteNotificationDriver")
	}
	if d.RemoteNotificationsAvailable() {
		t.Fatal("remote notifications available without the background mode")
	}
	errs := make(chan error, 1)
	d.RegisterForRemoteNotifications(func(_ string, err error) { errs <- err })
	select {
	case err := <-errs:
		if err == nil {
			t.Fatal("registering without the background mode must fail")
		}
	case <-time.After(time.Second):
		t.Fatal("registration never answered")
	}
}
