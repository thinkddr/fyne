// SPDX-License-Identifier: BSD-3-Clause

package gl

import (
	"errors"
	"image"
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type unsupportedRectanglePainter struct{}

func (unsupportedRectanglePainter) PaintSoftware(*image.NRGBA, canvas.RectanglePaintContext) error {
	panic("GL must not call a software provider")
}

func TestSuppliedRectangleGLRejectsWithoutFallback(t *testing.T) {
	rectangle := canvas.NewRectangle(color.Black)
	rectangle.Painter = unsupportedRectanglePainter{}
	p := &painter{}
	p.drawRectangle(rectangle, fyne.Position{}, fyne.NewSize(30, 20))
	if !errors.Is(rectangle.PaintError(), canvas.ErrRectanglePainterUnsupportedGL) {
		t.Fatalf("GL lost supplied-painter error: %v", rectangle.PaintError())
	}
	_ = rectangle.MinSize()
	if !errors.Is(rectangle.PaintError(), canvas.ErrRectanglePainterUnsupportedGL) {
		t.Fatal("measurement hid GL paint error")
	}
	rectangle.SetPaintError(nil)
	if rectangle.PaintError() != nil {
		t.Fatal("explicit paint error reset failed")
	}
}
