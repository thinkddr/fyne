package test

import (
	"fyne.io/fyne/v2"
	fynedriver "fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/software"
)

var dummyCanvas WindowlessCanvas

// WindowlessCanvas provides functionality for a canvas to operate without a window
type WindowlessCanvas interface {
	fyne.Canvas

	Padded() bool
	Resize(fyne.Size)
	SetPadded(bool)
	SetScale(float32)
}

// Canvas returns a reusable in-memory canvas used for testing
func Canvas() fyne.Canvas {
	if dummyCanvas == nil {
		dummyCanvas = NewCanvas()
	}

	return dummyCanvas
}

// NewCanvas returns a single use in-memory canvas used for testing.
// This canvas has no painter so calls to Capture() will return a blank image.
func NewCanvas() WindowlessCanvas {
	return wrapCanvas(software.NewCanvasWithPainter(nil))
}

// NewCanvasWithPainter allows creation of an in-memory canvas with a specific painter.
// The painter will be used to render in the Capture() call.
func NewCanvasWithPainter(painter fynedriver.Painter) WindowlessCanvas {
	return wrapCanvas(software.NewCanvasWithPainter(painter))
}

// NewTransparentCanvasWithPainter allows creation of an in-memory canvas with a specific painter without a background color.
// The painter will be used to render in the Capture() call.
//
// Since: 2.2
func NewTransparentCanvasWithPainter(painter fynedriver.Painter) WindowlessCanvas {
	return wrapCanvas(software.NewTransparentCanvasWithPainter(painter))
}

func wrapCanvas(c WindowlessCanvas) *canvas {
	return &canvas{WindowlessCanvas: c}
}

type canvas struct {
	WindowlessCanvas
	hovered desktop.Hoverable
}
