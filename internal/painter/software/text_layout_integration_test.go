package software_test

import (
	"bytes"
	"errors"
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

type painterTextLayout struct {
	t     *testing.T
	scale float32
	calls int
}

func (*painterTextLayout) Metrics() (fyne.Size, float32, error) {
	return fyne.NewSize(6, 4), 3, nil
}

func (l *painterTextLayout) PaintSoftware(dst *image.NRGBA, context canvas.TextPaintContext) error {
	l.calls++
	clip := image.Rect(int(10*l.scale), int(9*l.scale), int(20*l.scale), int(17*l.scale))
	if context.Clip != clip || dst.Bounds() != clip || dst.Stride != int(60*l.scale)*4 {
		l.t.Fatalf("actual painter clip/stride = %v, %v, %d", context.Clip, dst.Bounds(), dst.Stride)
	}
	if context.Origin != fyne.NewPos(10.25, 10.5) || context.Scale != l.scale {
		l.t.Fatalf("nested line origin/scale changed: %+v", context)
	}
	for y := clip.Min.Y; y < clip.Max.Y; y++ {
		for x := clip.Min.X; x < clip.Max.X; x++ {
			if got := dst.NRGBAAt(x, y); got != (color.NRGBA{R: 31, G: 79, B: 127, A: 255}) {
				l.t.Fatalf("real painter backdrop unavailable: %v", got)
			}
			dst.SetNRGBA(x, y, color.NRGBA{R: 201, G: 17, B: 83, A: 255})
		}
	}
	return nil
}

type textLayoutClip struct {
	widget.BaseWidget
	content fyne.CanvasObject
}

func (c *textLayoutClip) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.content)
}

func (*textLayoutClip) Scrolled(*fyne.ScrollEvent) {}

func TestPainterTextLayoutUsesAccumulatedDestinationAndAncestorClip(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		layout := &painterTextLayout{t: t, scale: scale}
		text := &canvas.Text{Layout: layout, Color: color.Black}
		text.Resize(fyne.NewSize(6, 4))
		text.Move(fyne.NewPos(0.25, 1.5))
		clip := &textLayoutClip{content: container.NewWithoutLayout(text)}
		clip.ExtendBaseWidget(clip)
		clip.Resize(fyne.NewSize(10, 8))
		clip.Move(fyne.NewPos(7, 5))
		nested := container.NewWithoutLayout(clip)
		nested.Move(fyne.NewPos(3, 4))
		background := canvas.NewRectangle(color.NRGBA{R: 31, G: 79, B: 127, A: 255})
		background.Resize(fyne.NewSize(60, 40))
		c := test.NewCanvas()
		c.SetPadded(false)
		c.SetScale(scale)
		c.SetContent(container.NewWithoutLayout(background, nested))
		c.Resize(fyne.NewSize(60, 40))
		painted := software.NewPainter().Paint(c).(*image.NRGBA)
		if layout.calls != 1 || text.TextLayoutError() != nil {
			t.Fatalf("actual tree did not paint supplied line once: %d, %v", layout.calls, text.TextLayoutError())
		}
		inside := image.Rect(int(10*scale), int(9*scale), int(20*scale), int(17*scale))
		for y := painted.Rect.Min.Y; y < painted.Rect.Max.Y; y++ {
			for x := painted.Rect.Min.X; x < painted.Rect.Max.X; x++ {
				want := color.NRGBA{R: 31, G: 79, B: 127, A: 255}
				if image.Pt(x, y).In(inside) {
					want = color.NRGBA{R: 201, G: 17, B: 83, A: 255}
				}
				if got := painted.NRGBAAt(x, y); got != want {
					t.Fatalf("scale %v pixel %d,%d = %v, want %v", scale, x, y, got, want)
				}
			}
		}
	}
}

type failingPainterTextLayout struct {
	size    fyne.Size
	failure error
}

func (l failingPainterTextLayout) Metrics() (fyne.Size, float32, error) {
	return l.size, 8, nil
}

func (l failingPainterTextLayout) PaintSoftware(*image.NRGBA, canvas.TextPaintContext) error {
	return l.failure
}

func TestPainterTextLayoutNilPreservesDefaultPainting(t *testing.T) {
	text := canvas.NewText("Default text", color.Black)
	c := test.NewCanvas()
	c.SetPadded(false)
	c.SetContent(text)
	c.Resize(fyne.NewSize(100, 40))
	p := software.NewPainter()
	before := p.Paint(c).(*image.NRGBA)
	if bytes.Equal(before.Pix, make([]byte, len(before.Pix))) || text.TextLayoutError() != nil {
		t.Fatal("nil layout did not use default text painting")
	}
	failure := errors.New("provider cannot paint this destination")
	text.Layout = failingPainterTextLayout{size: text.MinSize(), failure: failure}
	failed := p.Paint(c).(*image.NRGBA)
	if !bytes.Equal(failed.Pix, make([]byte, len(failed.Pix))) || !errors.Is(text.TextLayoutError(), failure) {
		t.Fatalf("supplied failure drew fallback glyphs or lost error: %v", text.TextLayoutError())
	}
	text.Layout = nil
	text.SetTextLayoutError(nil)
	after := p.Paint(c).(*image.NRGBA)
	if !bytes.Equal(before.Pix, after.Pix) || text.TextLayoutError() != nil {
		t.Fatalf("removing layout changed the default path: %v", text.TextLayoutError())
	}
}
