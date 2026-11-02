//go:build !windows || !ci

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package mobile

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTouchObserverContinuesAfterFilterPassBeforeOrdinaryDelivery(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	var filtered, observed, delivered []mobile.FilterPhase
	c.SetTouchFilter(func(e mobile.FilterEvent) mobile.FilterResult {
		filtered = append(filtered, e.Phase)
		return mobile.FilterPass
	})
	c.SetTouchObserver(func(e mobile.FilterEvent, result mobile.FilterResult) {
		assert.Equal(t, mobile.FilterPass, result)
		observed = append(observed, e.Phase)
	})
	for _, phase := range []mobile.FilterPhase{mobile.FilterBegin, mobile.FilterMove, mobile.FilterEnd} {
		c.filterTouch(filterEvent(phase, 0, 10, 10), func(e mobile.FilterEvent) {
			require.Len(t, observed, len(delivered)+1, "observation precedes ordinary dispatch")
			assert.Equal(t, e.Phase, observed[len(observed)-1])
			delivered = append(delivered, e.Phase)
		})
	}
	assert.Equal(t, []mobile.FilterPhase{mobile.FilterBegin}, filtered, "the filter keeps its Pass contract")
	assert.Equal(t, delivered, observed)
	assert.Empty(t, c.touchFilter.observed)
}

func TestTouchObserverDoesNotObserveBufferedReplayTwice(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	var observed, delivered []mobile.FilterPhase
	var results []mobile.FilterResult
	c.SetTouchFilter(func(e mobile.FilterEvent) mobile.FilterResult {
		if e.Phase == mobile.FilterEnd {
			return mobile.FilterPass
		}
		return mobile.FilterDefer
	})
	c.SetTouchObserver(func(e mobile.FilterEvent, result mobile.FilterResult) {
		observed = append(observed, e.Phase)
		results = append(results, result)
	})
	for _, phase := range []mobile.FilterPhase{mobile.FilterBegin, mobile.FilterMove, mobile.FilterEnd} {
		c.filterTouch(filterEvent(phase, 0, 10, 10), func(e mobile.FilterEvent) {
			delivered = append(delivered, e.Phase)
		})
	}
	assert.Equal(t, delivered, observed)
	assert.Equal(t, []mobile.FilterResult{mobile.FilterDefer, mobile.FilterDefer, mobile.FilterPass}, results)
}

func TestTouchObserverAlonePreservesMultiTouchAndOrphanEnds(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	var observed, delivered []int
	c.SetTouchObserver(func(e mobile.FilterEvent, result mobile.FilterResult) {
		assert.Equal(t, mobile.FilterPass, result)
		observed = append(observed, e.ID)
	})
	deliver := func(e mobile.FilterEvent) { delivered = append(delivered, e.ID) }
	for _, e := range []mobile.FilterEvent{
		filterEvent(mobile.FilterBegin, 0, 10, 10), filterEvent(mobile.FilterBegin, 1, 20, 20),
		filterEvent(mobile.FilterMove, 1, 25, 25), filterEvent(mobile.FilterEnd, 0, 10, 10),
		filterEvent(mobile.FilterCancel, 1, 25, 25),
	} {
		c.filterTouch(e, deliver)
	}
	assert.Equal(t, []int{0, 1, 1, 0, 1}, delivered)
	assert.Equal(t, delivered, observed)
	c.filterTouch(filterEvent(mobile.FilterEnd, 9, 0, 0), deliver)
	assert.Len(t, delivered, 5)
	assert.Len(t, observed, 5)
	assert.Empty(t, c.touchFilter.active)
	assert.Empty(t, c.touchFilter.observed)
}

type observerDrag struct {
	*widget.Label
	cancelled bool
	ends      int
	commits   int
}

func (d *observerDrag) Dragged(*fyne.DragEvent) {}

func (d *observerDrag) DragEnd() {
	d.ends++
	if !d.cancelled {
		d.commits++
	}
}

func TestTouchObserverDriverCancellationPrecedesDragEndAfterPass(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	item := &observerDrag{Label: widget.NewLabel("drag")}
	c.SetContent(item)
	c.Resize(fyne.NewSize(100, 100))
	c.SetTouchFilter(func(mobile.FilterEvent) mobile.FilterResult { return mobile.FilterPass })
	var observed []mobile.FilterPhase
	c.SetTouchObserver(func(e mobile.FilterEvent, _ mobile.FilterResult) {
		observed = append(observed, e.Phase)
		if e.Phase == mobile.FilterCancel {
			item.cancelled = true
		}
	})
	w := &window{canvas: c}
	d.touchCanvas(w, 10*c.Scale(), (10-tapYOffset)*c.Scale(), 0, mobile.FilterBegin)
	c.dragging = item
	d.touchCanvas(w, 20*c.Scale(), (20-tapYOffset)*c.Scale(), 0, mobile.FilterCancel)
	assert.Equal(t, []mobile.FilterPhase{mobile.FilterBegin, mobile.FilterCancel}, observed)
	assert.Equal(t, 1, item.ends)
	assert.Zero(t, item.commits, "OS cancellation must restore the drag before the driver ends it")
}

func TestTouchObserverReplacementAndRootChangesCancelPassedTouches(t *testing.T) {
	for _, change := range []string{"root", "overlay", "filter", "observer", "resize"} {
		t.Run(change, func(t *testing.T) {
			c := newCanvas(fyne.CurrentDevice()).(*canvas)
			var phases []mobile.FilterPhase
			c.SetTouchFilter(func(mobile.FilterEvent) mobile.FilterResult { return mobile.FilterPass })
			c.SetTouchObserver(func(e mobile.FilterEvent, _ mobile.FilterResult) {
				phases = append(phases, e.Phase)
			})
			c.filterTouch(filterEvent(mobile.FilterBegin, 0, 10, 10), func(mobile.FilterEvent) {})
			switch change {
			case "root":
				c.SetContent(widget.NewLabel("replacement"))
			case "overlay":
				c.Overlays().Add(widget.NewLabel("modal"))
			case "filter":
				c.SetTouchFilter(nil)
			case "observer":
				c.SetTouchObserver(nil)
			case "resize":
				c.Resize(fyne.NewSize(200, 100))
			}
			c.filterTouch(filterEvent(mobile.FilterEnd, 0, 10, 10), func(mobile.FilterEvent) {})
			assert.Equal(t, []mobile.FilterPhase{mobile.FilterBegin, mobile.FilterCancel}, phases)
		})
	}
}
