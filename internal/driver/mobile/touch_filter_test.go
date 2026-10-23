//go:build !windows || !ci

// Copyright 2026 Sytue. All rights reserved.
// Use of this source code is governed by the BSD-style license in LICENSE.

package mobile

import (
	"image/color"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func filterEvent(phase mobile.FilterPhase, id int, x, y float32) mobile.FilterEvent {
	return mobile.FilterEvent{Phase: phase, ID: id, Position: fyne.NewPos(x, y), Time: time.Now(), Size: fyne.NewSize(100, 100)}
}

func TestTouchFilterDeferredTapReplaysOriginalEvents(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	var got []mobile.FilterEvent
	c.SetTouchFilter(func(mobile.FilterEvent) mobile.FilterResult { return mobile.FilterDefer })
	begin := filterEvent(mobile.FilterBegin, 0, 10, 10)
	end := filterEvent(mobile.FilterEnd, 0, 11, 10)
	deliver := func(e mobile.FilterEvent) { got = append(got, e) }
	c.filterTouch(begin, deliver)
	assert.Empty(t, got, "pending recognition cannot clear focus or touch a child")
	c.filterTouch(end, deliver)
	require.Len(t, got, 2)
	assert.Equal(t, begin.Time, got[0].Time, "buffering does not restart long-press timing")
	assert.Equal(t, end.Time, got[1].Time)
	assert.Equal(t, []mobile.FilterPhase{mobile.FilterBegin, mobile.FilterEnd}, []mobile.FilterPhase{got[0].Phase, got[1].Phase})
	assert.Empty(t, c.touchFilter.active)
}

func TestTouchFilterRejectReplaysMovesOnceThenPassesThrough(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	var phases, delivered []mobile.FilterPhase
	c.SetTouchFilter(func(e mobile.FilterEvent) mobile.FilterResult {
		phases = append(phases, e.Phase)
		if e.Phase == mobile.FilterMove {
			return mobile.FilterPass
		}
		return mobile.FilterDefer
	})
	for _, phase := range []mobile.FilterPhase{mobile.FilterBegin, mobile.FilterMove, mobile.FilterMove, mobile.FilterEnd} {
		c.filterTouch(filterEvent(phase, 0, 10, 10), func(e mobile.FilterEvent) { delivered = append(delivered, e.Phase) })
	}
	assert.Equal(t, []mobile.FilterPhase{mobile.FilterBegin, mobile.FilterMove}, phases)
	assert.Equal(t, []mobile.FilterPhase{mobile.FilterBegin, mobile.FilterMove, mobile.FilterMove, mobile.FilterEnd}, delivered)
}

func TestTouchFilterClaimNeverTapsOrScrollsChild(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	var phases []mobile.FilterPhase
	c.SetTouchFilter(func(e mobile.FilterEvent) mobile.FilterResult {
		phases = append(phases, e.Phase)
		if e.Phase == mobile.FilterMove {
			return mobile.FilterClaim
		}
		return mobile.FilterDefer
	})
	for _, phase := range []mobile.FilterPhase{mobile.FilterBegin, mobile.FilterMove, mobile.FilterEnd} {
		c.filterTouch(filterEvent(phase, 0, 10, 10), func(mobile.FilterEvent) { t.Fatal("claimed gesture reached child") })
	}
	assert.Equal(t, []mobile.FilterPhase{mobile.FilterBegin, mobile.FilterMove, mobile.FilterEnd}, phases)
}

func TestTouchFilterSecondFingerPreservesUnclaimedMultiTouch(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	var received, delivered []mobile.FilterEvent
	c.SetTouchFilter(func(e mobile.FilterEvent) mobile.FilterResult {
		received = append(received, e)
		return mobile.FilterDefer
	})
	for _, e := range []mobile.FilterEvent{filterEvent(mobile.FilterBegin, 0, 10, 10), filterEvent(mobile.FilterMove, 0, 11, 10),
		filterEvent(mobile.FilterBegin, 1, 20, 20), filterEvent(mobile.FilterMove, 1, 21, 21), filterEvent(mobile.FilterEnd, 0, 11, 10), filterEvent(mobile.FilterEnd, 1, 21, 21)} {
		c.filterTouch(e, func(e mobile.FilterEvent) { delivered = append(delivered, e) })
	}
	require.Len(t, received, 3)
	assert.Equal(t, mobile.FilterCancel, received[2].Phase)
	require.Len(t, delivered, 6)
	assert.Equal(t, []int{0, 0, 1, 1, 0, 1}, []int{delivered[0].ID, delivered[1].ID, delivered[2].ID, delivered[3].ID, delivered[4].ID, delivered[5].ID})
	assert.Empty(t, c.touchFilter.active)
}

func TestTouchFilterClaimedMultiTouchCancelsUntilAllFingersLift(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	var phases []mobile.FilterPhase
	c.SetTouchFilter(func(e mobile.FilterEvent) mobile.FilterResult {
		phases = append(phases, e.Phase)
		return mobile.FilterClaim
	})
	for _, e := range []mobile.FilterEvent{filterEvent(mobile.FilterBegin, 0, 10, 10), filterEvent(mobile.FilterBegin, 1, 20, 20),
		filterEvent(mobile.FilterBegin, 2, 30, 30), filterEvent(mobile.FilterEnd, 0, 10, 10), filterEvent(mobile.FilterEnd, 1, 20, 20), filterEvent(mobile.FilterCancel, 2, 30, 30)} {
		c.filterTouch(e, func(mobile.FilterEvent) { t.Fatal("cancelled claim leaked to ordinary widgets") })
	}
	assert.Equal(t, []mobile.FilterPhase{mobile.FilterBegin, mobile.FilterCancel}, phases)
	assert.Empty(t, c.touchFilter.active)
	c.filterTouch(filterEvent(mobile.FilterEnd, 9, 10, 10), func(mobile.FilterEvent) { t.Fatal("orphan end created a tap") })
	c.filterTouch(filterEvent(mobile.FilterBegin, 0, 10, 10), func(mobile.FilterEvent) {})
	assert.Equal(t, mobile.FilterBegin, phases[len(phases)-1], "new gesture works after all fingers leave")
}

func TestTouchFilterRootAndOverlayChangesCancelWithoutReplay(t *testing.T) {
	for _, change := range []string{"root", "overlay", "filter", "resize"} {
		t.Run(change, func(t *testing.T) {
			c := newCanvas(fyne.CurrentDevice()).(*canvas)
			var phases []mobile.FilterPhase
			c.SetTouchFilter(func(e mobile.FilterEvent) mobile.FilterResult {
				phases = append(phases, e.Phase)
				return mobile.FilterDefer
			})
			deliver := func(mobile.FilterEvent) { t.Fatal("obsolete candidate reached replacement content") }
			c.filterTouch(filterEvent(mobile.FilterBegin, 0, 10, 10), deliver)
			switch change {
			case "root":
				c.SetContent(widget.NewLabel("replacement"))
			case "overlay":
				c.Overlays().Add(widget.NewLabel("modal"))
			case "filter":
				c.SetTouchFilter(nil)
			case "resize":
				c.Resize(fyne.NewSize(200, 100))
			}
			c.filterTouch(filterEvent(mobile.FilterEnd, 0, 10, 10), deliver)
			assert.Equal(t, []mobile.FilterPhase{mobile.FilterBegin, mobile.FilterCancel}, phases)
		})
	}
}

func TestTouchFilterHitPathHonoursClippingAndOverlay(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	child := canvas.NewRectangle(color.White)
	child.SetMinSize(fyne.NewSize(200, 200))
	scroll := container.NewVScroll(child)
	c.padded = false
	c.SetContent(container.NewWithoutLayout(scroll))
	c.Resize(fyne.NewSize(100, 100))
	scroll.Resize(fyne.NewSize(40, 40))
	path := c.touchPath(fyne.NewPos(10, 10))
	assert.Contains(t, path, scroll)
	assert.Contains(t, path, child)
	assert.NotContains(t, c.touchPath(fyne.NewPos(80, 80)), child, "children outside the clipped scroll cannot nominate refresh")
	overlay := widget.NewLabel("overlay")
	c.Overlays().Add(overlay)
	path = c.touchPath(fyne.NewPos(10, 10))
	assert.Contains(t, path, overlay)
	assert.NotContains(t, path, scroll, "a modal cannot nominate the list underneath")
}

func TestTouchFilterBoundedBufferAndNativeCancellation(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	var cancelled, count int
	c.SetTouchFilter(func(e mobile.FilterEvent) mobile.FilterResult {
		if e.Phase == mobile.FilterCancel {
			cancelled++
		}
		return mobile.FilterDefer
	})
	deliver := func(mobile.FilterEvent) { count++ }
	c.filterTouch(filterEvent(mobile.FilterBegin, 0, 10, 10), deliver)
	for i := 0; i < 128; i++ {
		c.filterTouch(filterEvent(mobile.FilterMove, 0, 10, 10), deliver)
	}
	assert.Equal(t, 129, count)
	assert.Equal(t, 1, cancelled)
	c.filterTouch(filterEvent(mobile.FilterCancel, 0, 10, 10), deliver)
	assert.Empty(t, c.touchFilter.active)
	assert.Empty(t, c.touchFilter.pending)
}

func TestTouchFilterDriverDispatchPreservesFocusAndCancellation(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	content := &touchableLabel{Label: widget.NewLabel("touch"), ids: make(map[int]bool)}
	content.ExtendBaseWidget(content)
	entry := widget.NewEntry()
	c.SetContent(container.NewWithoutLayout(entry, content))
	c.Resize(fyne.NewSize(100, 100))
	entry.Resize(fyne.NewSize(30, 30))
	content.Resize(fyne.NewSize(30, 30))
	content.Move(fyne.NewPos(50, 50))
	c.Focus(entry)
	w := &window{canvas: c}
	c.SetTouchFilter(func(e mobile.FilterEvent) mobile.FilterResult {
		if e.Phase == mobile.FilterMove {
			return mobile.FilterClaim
		}
		return mobile.FilterDefer
	})
	d.touchCanvas(w, 55*c.Scale(), (55-tapYOffset)*c.Scale(), 0, mobile.FilterBegin)
	assert.False(t, content.down)
	assert.Equal(t, entry, c.Focused(), "recognition does not dismiss the keyboard")
	d.touchCanvas(w, 70*c.Scale(), (55-tapYOffset)*c.Scale(), 0, mobile.FilterMove)
	d.touchCanvas(w, 70*c.Scale(), (55-tapYOffset)*c.Scale(), 0, mobile.FilterEnd)
	assert.False(t, content.down)
	assert.False(t, content.up)
	assert.Equal(t, entry, c.Focused())

	c.SetTouchFilter(func(mobile.FilterEvent) mobile.FilterResult { return mobile.FilterPass })
	d.touchCanvas(w, 55*c.Scale(), (55-tapYOffset)*c.Scale(), 0, mobile.FilterBegin)
	assert.True(t, content.down)
	d.touchCanvas(w, 55*c.Scale(), (55-tapYOffset)*c.Scale(), 0, mobile.FilterCancel)
	assert.True(t, content.cancel)
	assert.False(t, content.up, "OS cancellation cannot become a tap")
	assert.Empty(t, c.lastTapDown)
}

func TestTouchFilterEndPathDoesNotReuseRemovedView(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	old := widget.NewLabel("old list")
	root := container.NewStack(old)
	c.SetContent(root)
	c.Resize(fyne.NewSize(100, 100))
	var beginPath, endPath []fyne.CanvasObject
	c.SetTouchFilter(func(e mobile.FilterEvent) mobile.FilterResult {
		switch e.Phase {
		case mobile.FilterBegin:
			beginPath = append([]fyne.CanvasObject(nil), e.Path...)
		case mobile.FilterEnd:
			endPath = append([]fyne.CanvasObject(nil), e.Path...)
			return mobile.FilterClaim // Discard the tap when its original list has gone.
		}
		return mobile.FilterDefer
	})
	deliver := func(mobile.FilterEvent) { t.Fatal("removed view gesture retargeted its replacement") }
	c.filterTouch(filterEvent(mobile.FilterBegin, 0, 10, 10), deliver)
	replacement := widget.NewLabel("new list")
	root.Objects = []fyne.CanvasObject{replacement}
	root.Refresh()
	c.filterTouch(filterEvent(mobile.FilterEnd, 0, 10, 10), deliver)
	assert.Contains(t, beginPath, old)
	assert.NotContains(t, endPath, old)
	assert.Contains(t, endPath, replacement)
}
