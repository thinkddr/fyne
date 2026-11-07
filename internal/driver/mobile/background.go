// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package mobile

import (
	"context"
	"time"

	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/internal/driver/mobile/app"
)

var _ mobile.BackgroundDriver = (*driver)(nil)

// LaunchedInBackground implements mobile.BackgroundDriver.
func (*driver) LaunchedInBackground() bool { return app.LaunchedInBackground() }

// OnFirstScene implements mobile.BackgroundDriver.
func (*driver) OnFirstScene(f func()) { app.OnFirstScene(f) }

// ProtectedDataAvailable implements mobile.BackgroundDriver.
func (*driver) ProtectedDataAvailable() bool { return app.ProtectedDataAvailable() }

// OnBackgroundRefresh implements mobile.BackgroundDriver.
func (*driver) OnBackgroundRefresh(every time.Duration, f func(ctx context.Context) (ok bool, retry time.Duration)) {
	app.OnBackgroundRefresh(every, f)
}
