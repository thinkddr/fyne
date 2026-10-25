// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Daniel Delgado Ramos

package main

/*
#cgo LDFLAGS: -framework UIKit -framework Foundation -framework GLKit
void geometryProbeUpdate(double width, double height, double scale, int taps, double tapX, double tapY);
*/
import "C"

import (
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type probe struct {
	widget.BaseWidget
	mu   sync.Mutex
	taps int
	last fyne.Position
}

func (p *probe) MinSize() fyne.Size { return fyne.NewSize(0, 0) }
func (p *probe) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.NRGBA{R: 30, G: 100, B: 160, A: 255}))
}
func (p *probe) Tapped(event *fyne.PointEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.taps++
	p.last = event.AbsolutePosition
}

func main() {
	a := app.NewWithID("io.fyne.geometry")
	w := a.NewWindow("UIKit geometry")
	w.SetPadded(false)
	p := &probe{}
	p.ExtendBaseWidget(p)
	w.SetContent(p)
	go func() {
		for range time.Tick(100 * time.Millisecond) {
			fyne.Do(func() {
				p.mu.Lock()
				taps, last := p.taps, p.last
				p.mu.Unlock()
				size := w.Canvas().Size()
				scale := a.Driver().Device().SystemScaleForWindow(w)
				C.geometryProbeUpdate(C.double(size.Width), C.double(size.Height), C.double(scale),
					C.int(taps), C.double(last.X), C.double(last.Y))
			})
		}
	}()
	w.ShowAndRun()
}
