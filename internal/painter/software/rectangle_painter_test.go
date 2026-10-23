// SPDX-License-Identifier: BSD-3-Clause

package software

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type suppliedRectangleFunc func(*image.NRGBA, canvas.RectanglePaintContext) error

func (f suppliedRectangleFunc) PaintSoftware(dst *image.NRGBA, context canvas.RectanglePaintContext) error {
	return f(dst, context)
}

func TestSuppliedRectangleUsesActualClipAndStride(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		base := image.NewNRGBA(image.Rect(-3, -2, 24, 18))
		background := color.NRGBA{R: 31, G: 67, B: 103, A: 197}
		for y := base.Rect.Min.Y; y < base.Rect.Max.Y; y++ {
			for x := base.Rect.Min.X; x < base.Rect.Max.X; x++ {
				base.SetNRGBA(x, y, background)
			}
		}
		clip := image.Rect(-9, 3, 10, 25)
		wantClip := image.Rect(-3, 3, 10, 18)
		calls := 0
		rectangle := canvas.NewRectangle(color.Black)
		rectangle.Resize(fyne.NewSize(4, 5))
		rectangle.Painter = suppliedRectangleFunc(func(dst *image.NRGBA, context canvas.RectanglePaintContext) error {
			calls++
			if context.Origin != fyne.NewPos(-1.25, 2.5) || context.Size != fyne.NewSize(4, 5) ||
				context.Scale != scale || context.Clip != wantClip || dst.Bounds() != wantClip || dst.Stride != base.Stride {
				t.Fatalf("changed geometry or packed destination: %+v, %v, %d", context, dst.Bounds(), dst.Stride)
			}
			for y := dst.Rect.Min.Y; y < dst.Rect.Max.Y; y++ {
				for x := dst.Rect.Min.X; x < dst.Rect.Max.X; x++ {
					if dst.NRGBAAt(x, y) != background {
						t.Fatal("accumulated backdrop was replaced")
					}
					dst.SetNRGBA(x, y, color.NRGBA{R: 201, A: 255})
				}
			}
			return nil
		})
		drawRectangle(textLayoutCanvas{scale: scale}, rectangle, fyne.NewPos(-1.25, 2.5), base, clip)
		if calls != 1 || rectangle.PaintError() != nil {
			t.Fatalf("provider calls/error: %d, %v", calls, rectangle.PaintError())
		}
		for y := base.Rect.Min.Y; y < base.Rect.Max.Y; y++ {
			for x := base.Rect.Min.X; x < base.Rect.Max.X; x++ {
				want := background
				if image.Pt(x, y).In(wantClip) {
					want = color.NRGBA{R: 201, A: 255}
				}
				if base.NRGBAAt(x, y) != want {
					t.Fatalf("unexpected pixel %d,%d", x, y)
				}
			}
		}
	}
}

func TestSuppliedRectangleRejectsGeometryBeforeWriting(t *testing.T) {
	for _, sample := range []struct {
		position fyne.Position
		size     fyne.Size
		scale    float32
	}{
		{fyne.NewPos(float32(math.NaN()), 0), fyne.NewSize(3, 4), 1},
		{fyne.NewPos(0, float32(math.Inf(1))), fyne.NewSize(3, 4), 1},
		{fyne.Position{}, fyne.NewSize(-1, 4), 1},
		{fyne.Position{}, fyne.NewSize(3, float32(math.NaN())), 1},
		{fyne.Position{}, fyne.NewSize(3, 4), 0},
		{fyne.Position{}, fyne.NewSize(3, 4), float32(math.Inf(1))},
	} {
		base := image.NewNRGBA(image.Rect(0, 0, 8, 8))
		rectangle := canvas.NewRectangle(color.Black)
		rectangle.Resize(sample.size)
		rectangle.Painter = suppliedRectangleFunc(func(*image.NRGBA, canvas.RectanglePaintContext) error {
			t.Fatal("invalid geometry reached provider")
			return nil
		})
		drawRectangle(textLayoutCanvas{scale: sample.scale}, rectangle, sample.position, base, base.Bounds())
		if !errors.Is(rectangle.PaintError(), canvas.ErrRectanglePainterInvalidGeometry) || !bytes.Equal(base.Pix, make([]byte, len(base.Pix))) {
			t.Fatalf("invalid geometry wrote pixels or lost error: %v", rectangle.PaintError())
		}
	}
}

func TestSuppliedRectangleFailureNeverPaintsDefaultAndCanRecover(t *testing.T) {
	base := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	rectangle := canvas.NewRectangle(color.Black)
	rectangle.Resize(fyne.NewSize(5, 5))
	failure := errors.New("unsupported native plan")
	rectangle.Painter = suppliedRectangleFunc(func(*image.NRGBA, canvas.RectanglePaintContext) error { return failure })
	c := textLayoutCanvas{scale: 1}
	drawRectangle(c, rectangle, fyne.Position{}, base, base.Bounds())
	if !errors.Is(rectangle.PaintError(), failure) || !bytes.Equal(base.Pix, make([]byte, len(base.Pix))) {
		t.Fatalf("provider failure fell back or lost error: %v", rectangle.PaintError())
	}
	_ = rectangle.MinSize()
	rectangle.Refresh()
	if !errors.Is(rectangle.PaintError(), failure) {
		t.Fatal("measurement or refresh hid paint failure")
	}
	rectangle.Painter = suppliedRectangleFunc(func(*image.NRGBA, canvas.RectanglePaintContext) error { return nil })
	drawRectangle(c, rectangle, fyne.Position{}, base, image.Rect(20, 20, 21, 21))
	if !errors.Is(rectangle.PaintError(), failure) {
		t.Fatal("a clipped paint hid the previous failure")
	}
	drawRectangle(c, rectangle, fyne.Position{}, base, base.Bounds())
	if rectangle.PaintError() != nil {
		t.Fatal("successful paint did not clear failure")
	}
	rectangle.Painter = nil
	drawRectangle(c, rectangle, fyne.Position{}, base, base.Bounds())
	if base.NRGBAAt(2, 2) != (color.NRGBA{A: 255}) {
		t.Fatal("nil provider changed default rectangle painting")
	}
}
