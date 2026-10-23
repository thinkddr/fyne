// Copyright 2026 Sytue. All rights reserved.
// Use of this source code is governed by the BSD-style license in LICENSE.

package mobile

import (
	"time"

	"fyne.io/fyne/v2"
)

// FilterPhase identifies a touch sequence transition.
type FilterPhase uint8

const (
	FilterBegin FilterPhase = iota
	FilterMove
	FilterEnd
	FilterCancel
)

// FilterEvent describes a touch in canvas coordinates. Path contains the visible
// hit object's ancestors, from root to leaf, for FilterBegin and FilterEnd. It respects
// clipping and the top overlay. Time is captured before any deferred dispatch.
type FilterEvent struct {
	Phase    FilterPhase
	ID       int
	Position fyne.Position
	Time     time.Time
	Size     fyne.Size
	Path     []fyne.CanvasObject
}

// FilterResult controls dispatch of a single-finger sequence.
type FilterResult uint8

const (
	// FilterPass delivers deferred events in order and leaves this sequence to
	// ordinary widgets. No further filter events arrive until all fingers lift.
	FilterPass FilterResult = iota
	// FilterDefer buffers events before they change focus or reach a child.
	// Return Pass or Claim as soon as the gesture's direction is known.
	FilterDefer
	// FilterClaim consumes a deferred sequence through its terminal event.
	// A claimed sequence cannot subsequently be returned to ordinary widgets.
	FilterClaim
)

// TouchFilterCanvas is an optional mobile canvas API. The callback runs on the
// UI goroutine and must not block. It must not modify or retain the Path slice.
// Replacing the filter, changing content/overlays or a second finger
// cancels recognition. Perform actions only on FilterEnd, never FilterCancel.
//
// A deferred sequence that ends without being claimed is delivered normally.
// A second finger replays deferred events and preserves ordinary multi-touch.
// After a claim, additional fingers cancel recognition and remain consumed
// until every finger lifts. A maximum of 128 events may be deferred.
type TouchFilterCanvas interface {
	SetTouchFilter(func(FilterEvent) FilterResult)
}
