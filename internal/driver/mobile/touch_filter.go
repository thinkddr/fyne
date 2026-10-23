// Copyright 2026 Sytue. All rights reserved.
// Use of this source code is governed by the BSD-style license in LICENSE.

package mobile

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/mobile"
	intdriver "fyne.io/fyne/v2/internal/driver"
)

type touchFilterState struct {
	callback func(mobile.FilterEvent) mobile.FilterResult
	active   map[int]bool
	pending  []mobile.FilterEvent
	last     mobile.FilterEvent
	mode     mobile.FilterResult
	blocked  bool
}

func (c *canvas) SetTouchFilter(callback func(mobile.FilterEvent) mobile.FilterResult) {
	c.cancelTouchFilter()
	c.touchFilter.callback = callback
}

func (c *canvas) cancelTouchFilter() {
	s := &c.touchFilter
	if s.blocked || (s.mode != mobile.FilterDefer && s.mode != mobile.FilterClaim) {
		return
	}
	e := s.last
	e.Phase, e.Path, e.Time = mobile.FilterCancel, nil, time.Now()
	callback := s.callback
	s.pending, s.mode, s.blocked = nil, mobile.FilterClaim, true
	if callback != nil {
		callback(e)
	}
}

func (c *canvas) filterTouch(e mobile.FilterEvent, deliver func(mobile.FilterEvent)) {
	s := &c.touchFilter
	if s.callback == nil && len(s.active) == 0 {
		deliver(e)
		return
	}
	if s.active == nil {
		s.active = make(map[int]bool)
	}
	if e.Phase == mobile.FilterBegin {
		if len(s.active) == 0 {
			s.mode, s.blocked = mobile.FilterDefer, false
		} else if s.mode == mobile.FilterDefer {
			// Cancel recognition before replay, but keep the pending events so
			// both fingers still reach the ordinary touch implementation.
			pending := s.pending
			c.cancelTouchFilter()
			s.mode = mobile.FilterPass
			for _, buffered := range pending {
				deliver(buffered)
			}
		} else if s.mode == mobile.FilterClaim {
			c.cancelTouchFilter()
		}
		s.active[e.ID] = true
	} else if !s.active[e.ID] {
		return // A terminal event must never create a new gesture or tap.
	}
	terminal := e.Phase == mobile.FilterEnd || e.Phase == mobile.FilterCancel
	defer func() {
		if terminal {
			delete(s.active, e.ID)
			if len(s.active) == 0 {
				s.pending, s.mode, s.blocked = nil, mobile.FilterPass, false
			}
		}
	}()
	if s.blocked || s.callback == nil || s.mode == mobile.FilterPass {
		if s.mode != mobile.FilterClaim {
			deliver(e)
		}
		return
	}
	if e.Phase == mobile.FilterBegin || e.Phase == mobile.FilterEnd {
		e.Path = c.touchPath(e.Position)
	}
	s.last = e
	result := s.callback(e)
	if s.blocked { // The callback may have replaced the root or opened an overlay.
		return
	}
	if s.mode == mobile.FilterClaim {
		return
	}
	if result == mobile.FilterClaim {
		s.pending, s.mode = nil, mobile.FilterClaim
		return
	}
	s.pending = append(s.pending, e)
	if result != mobile.FilterDefer || terminal || len(s.pending) >= 128 {
		pending := s.pending
		s.pending, s.mode = nil, mobile.FilterPass
		if result == mobile.FilterDefer && !terminal {
			// An indecisive filter must not retain an unbounded event stream.
			e.Phase, e.Path = mobile.FilterCancel, nil
			s.callback(e)
		}
		for _, buffered := range pending {
			deliver(buffered)
		}
	}
}

func (c *canvas) touchPath(pos fyne.Position) []fyne.CanvasObject {
	roots := []fyne.CanvasObject{c.windowHead, c.content}
	if c.menu != nil {
		roots = []fyne.CanvasObject{c.menu}
	}
	if top := c.Overlays().Top(); top != nil {
		roots = []fyne.CanvasObject{top}
	}
	for _, root := range roots {
		var stack, found []fyne.CanvasObject
		intdriver.WalkVisibleObjectTree(root, func(o fyne.CanvasObject, p, cp fyne.Position, cs fyne.Size) bool {
			stack = append(stack, o)
			if pos.X >= p.X && pos.Y >= p.Y && pos.X < p.X+o.Size().Width && pos.Y < p.Y+o.Size().Height &&
				pos.X >= cp.X && pos.Y >= cp.Y && pos.X < cp.X+cs.Width && pos.Y < cp.Y+cs.Height {
				found = append(found[:0], stack...)
			}
			return false
		}, func(fyne.CanvasObject, fyne.Position, fyne.CanvasObject) {
			stack = stack[:len(stack)-1]
		})
		if len(found) != 0 {
			return found
		}
	}
	return nil
}

func (c *canvas) tapCancel(pos fyne.Position, tapID int) {
	if touched := c.touched[tapID]; touched != nil {
		relative := pos
		if object, ok := touched.(fyne.CanvasObject); ok {
			absolute := intdriver.AbsolutePositionForObject(object, []fyne.CanvasObject{c.Overlays().Top(), c.menu, c.windowHead, c.content})
			relative = pos.Subtract(absolute)
		}
		touched.TouchCancel(&mobile.TouchEvent{PointEvent: fyne.PointEvent{Position: relative, AbsolutePosition: pos}, ID: tapID})
		delete(c.touched, tapID)
	}
	if tapID == 0 && c.dragging != nil {
		c.dragging.DragEnd()
		if c.draggingOuter != nil {
			c.draggingOuter.DragEnd()
		}
		c.dragging, c.draggingOuter = nil, nil
		c.otherDirection = container.ScrollBoth
	}
	delete(c.lastTapDown, tapID)
	delete(c.lastTapDownPos, tapID)
	delete(c.lastTapDelta, tapID)
}
