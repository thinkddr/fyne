// SPDX-License-Identifier: BSD-3-Clause

package software_test

import (
	"image"
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/internal/painter/software"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

type rectanglePaintPlan struct {
	t     *testing.T
	scale float32
	calls int
}

func (p *rectanglePaintPlan) PaintSoftware(dst *image.NRGBA, context canvas.RectanglePaintContext) error {
	p.calls++
	clip := image.Rect(int(10*p.scale), int(9*p.scale), int(20*p.scale), int(17*p.scale))
	if context.Clip != clip || dst.Bounds() != clip || dst.Stride != int(60*p.scale)*4 ||
		context.Origin != fyne.NewPos(10.25, 10.5) || context.Size != fyne.NewSize(3, 2) || context.Scale != p.scale {
		p.t.Fatalf("actual tree changed paint context/stride: %+v, %v, %d", context, dst.Bounds(), dst.Stride)
	}
	for y := clip.Min.Y; y < clip.Max.Y; y++ {
		for x := clip.Min.X; x < clip.Max.X; x++ {
			if dst.NRGBAAt(x, y) != (color.NRGBA{R: 31, G: 79, B: 127, A: 255}) {
				p.t.Fatal("provider did not receive earlier canvas painting")
			}
			dst.SetNRGBA(x, y, color.NRGBA{R: 201, G: 17, B: 83, A: 255})
		}
	}
	return nil
}

type rectanglePaintClip struct {
	widget.BaseWidget
	content fyne.CanvasObject
}

func (c *rectanglePaintClip) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.content)
}

func (*rectanglePaintClip) Scrolled(*fyne.ScrollEvent) {}

func TestPainterSuppliedRectangleUsesCanvasBackdropAncestorClipAndOrder(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		plan := &rectanglePaintPlan{t: t, scale: scale}
		rectangle := canvas.NewRectangle(color.Black)
		rectangle.Painter = plan
		rectangle.Resize(fyne.NewSize(3, 2))
		rectangle.Move(fyne.NewPos(0.25, 1.5))
		clip := &rectanglePaintClip{content: container.NewWithoutLayout(rectangle)}
		clip.ExtendBaseWidget(clip)
		clip.Resize(fyne.NewSize(10, 8))
		clip.Move(fyne.NewPos(7, 5))
		nested := container.NewWithoutLayout(clip)
		nested.Move(fyne.NewPos(3, 4))
		background := canvas.NewRectangle(color.NRGBA{R: 31, G: 79, B: 127, A: 255})
		background.Resize(fyne.NewSize(60, 40))
		later := canvas.NewRectangle(color.NRGBA{G: 211, A: 255})
		later.Resize(fyne.NewSize(2, 2))
		later.Move(fyne.NewPos(18, 15))
		c := test.NewCanvas()
		c.SetPadded(false)
		c.SetScale(scale)
		c.SetContent(container.NewWithoutLayout(background, nested, later))
		c.Resize(fyne.NewSize(60, 40))
		painted := software.NewPainter().Paint(c).(*image.NRGBA)
		if plan.calls != 1 || rectangle.PaintError() != nil {
			t.Fatalf("actual tree calls/error: %d, %v", plan.calls, rectangle.PaintError())
		}
		inside := image.Rect(int(10*scale), int(9*scale), int(20*scale), int(17*scale))
		laterBox := image.Rect(int(18*scale), int(15*scale), int(20*scale), int(17*scale))
		for y := painted.Rect.Min.Y; y < painted.Rect.Max.Y; y++ {
			for x := painted.Rect.Min.X; x < painted.Rect.Max.X; x++ {
				want := color.NRGBA{R: 31, G: 79, B: 127, A: 255}
				if image.Pt(x, y).In(inside) {
					want = color.NRGBA{R: 201, G: 17, B: 83, A: 255}
				}
				if image.Pt(x, y).In(laterBox) {
					want = color.NRGBA{G: 211, A: 255}
				}
				if painted.NRGBAAt(x, y) != want {
					t.Fatalf("scale %v pixel %d,%d changed paint order or clip", scale, x, y)
				}
			}
		}
	}
}
