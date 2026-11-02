// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

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

// TouchObserverCanvas optionally observes touch lifecycle without deciding dispatch.
// The callback runs on the UI goroutine once per incoming event, before delivery to
// ordinary widgets, including after a filter returns FilterPass. The result describes
// whether the event is delivered, deferred or consumed; the observer cannot change it.
// Replayed events are not observed twice. Path is populated for Begin and End and must
// not be modified or retained. Observers must not block or replace content/overlays.
// Replacing content, overlays, filter or observer cancels active observations before
// ordinary drag completion. Cancel must restore visual state, never commit an action.
type TouchObserverCanvas interface {
	SetTouchObserver(func(FilterEvent, FilterResult))
}
