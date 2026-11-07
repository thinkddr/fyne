// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package mobile

import (
	"testing"

	"fyne.io/fyne/v2/driver/mobile"
)

// Apps reach the background API by asserting the driver they were given.
func TestTheDriverOffersTheBackgroundAPI(t *testing.T) {
	d, ok := NewGoMobileDriver().(mobile.BackgroundDriver)
	if !ok {
		t.Fatal("the mobile driver does not implement mobile.BackgroundDriver")
	}
	ran := false
	d.OnFirstScene(func() { ran = true })
	if !ran || d.LaunchedInBackground() || !d.ProtectedDataAvailable() {
		t.Fatal("without scenes the first scene is now, with protected data and no background launch")
	}
}
