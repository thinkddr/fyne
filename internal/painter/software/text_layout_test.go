package software

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type textLayoutCanvas struct {
	fyne.Canvas
	scale float32
}

func (c textLayoutCanvas) Scale() float32 { return c.scale }

type softwareTextLayout struct {
	size        fyne.Size
	baseline    float32
	metricError error
	paint       func(*image.NRGBA, canvas.TextPaintContext) error
}

func (l *softwareTextLayout) Metrics() (fyne.Size, float32, error) {
	return l.size, l.baseline, l.metricError
}

func (l *softwareTextLayout) PaintSoftware(dst *image.NRGBA, context canvas.TextPaintContext) error {
	return l.paint(dst, context)
}

func TestTextLayoutDestinationAndPlacement(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		for _, tc := range []struct {
			name      string
			alignment fyne.TextAlign
			x         float32
		}{
			{"leading", fyne.TextAlignLeading, 3.25},
			{"center", fyne.TextAlignCenter, 6.25},
			{"trailing", fyne.TextAlignTrailing, 9.25},
		} {
			t.Run(fmt.Sprintf("%s/scale-%g", tc.name, scale), func(t *testing.T) {
				base := image.NewNRGBA(image.Rect(10, 20, 31, 33))
				background := color.NRGBA{R: 23, G: 71, B: 113, A: 197}
				draw.Draw(base, base.Bounds(), image.NewUniform(background), image.Point{}, draw.Src)
				before := append([]byte(nil), base.Pix...)
				requestedClip := image.Rect(8, 18, 24, 30)
				clip := image.Rect(10, 20, 24, 30)
				ink := color.NRGBA{R: 211, G: 53, B: 97, A: 128}
				painted := color.NRGBA{R: 24, G: 72, B: 114, A: 198}
				calls := 0
				layout := &softwareTextLayout{size: fyne.NewSize(6.5, 4), baseline: 3}
				layout.paint = func(dst *image.NRGBA, context canvas.TextPaintContext) error {
					calls++
					if dst.Bounds() != clip || dst.Stride != base.Stride || context.Clip != clip {
						t.Fatalf("destination geometry = %v stride %d, context %v", dst.Bounds(), dst.Stride, context)
					}
					if context.Origin != fyne.NewPos(tc.x, 7.5) || context.Scale != scale || context.Color != ink {
						t.Fatalf("paint context changed logical placement/color: %+v", context)
					}
					for y := clip.Min.Y; y < clip.Max.Y; y++ {
						for x := clip.Min.X; x < clip.Max.X; x++ {
							if got := dst.NRGBAAt(x, y); got != background {
								t.Fatalf("provider did not receive accumulated backdrop at %d,%d: %v", x, y, got)
							}
							dst.SetNRGBA(x, y, painted)
						}
					}
					return nil
				}
				text := &canvas.Text{Text: "", Layout: layout, Color: ink, Alignment: tc.alignment}
				text.Resize(fyne.NewSize(12.5, 10))
				drawText(textLayoutCanvas{scale: scale}, text, fyne.NewPos(3.25, 4.5), base, requestedClip)
				if calls != 1 || text.TextLayoutError() != nil {
					t.Fatalf("supplied empty line not painted exactly once: %d, %v", calls, text.TextLayoutError())
				}
				for y := base.Rect.Min.Y; y < base.Rect.Max.Y; y++ {
					for x := base.Rect.Min.X; x < base.Rect.Max.X; x++ {
						if image.Pt(x, y).In(clip) {
							if got := base.NRGBAAt(x, y); got != painted {
								t.Fatalf("provider writes did not reach real buffer: %d,%d = %v", x, y, got)
							}
							continue
						}
						offset := base.PixOffset(x, y)
						if !bytes.Equal(base.Pix[offset:offset+4], before[offset:offset+4]) {
							t.Fatalf("pixels outside clip changed at %d,%d", x, y)
						}
					}
				}
			})
		}
	}
}

func TestTextLayoutPaintErrorSurvivesMetricsAndClearsOnPaint(t *testing.T) {
	base := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	before := append([]byte(nil), base.Pix...)
	failure := errors.New("paint adapter is closed")
	layout := &softwareTextLayout{
		size: fyne.NewSize(10, 10), baseline: 8,
		paint: func(*image.NRGBA, canvas.TextPaintContext) error {
			return failure
		},
	}
	text := &canvas.Text{Text: "default glyphs must not appear", Color: color.Black, Layout: layout}
	drawText(textLayoutCanvas{scale: 1}, text, fyne.Position{}, base, base.Bounds())
	if !bytes.Equal(base.Pix, before) || !errors.Is(text.TextLayoutError(), failure) {
		t.Fatalf("failed paint changed pixels or hid error: %v", text.TextLayoutError())
	}
	text.MinSize()
	if !errors.Is(text.TextLayoutError(), failure) {
		t.Fatalf("MinSize cleared paint error: %v", text.TextLayoutError())
	}
	layout.paint = func(dst *image.NRGBA, context canvas.TextPaintContext) error {
		dst.SetNRGBA(2, 3, color.NRGBA{R: 41, A: 255})
		return nil
	}
	drawText(textLayoutCanvas{scale: 1}, text, fyne.Position{}, base, base.Bounds())
	if text.TextLayoutError() != nil || base.NRGBAAt(2, 3).R != 41 {
		t.Fatalf("successful paint did not recover: %v", text.TextLayoutError())
	}
}

func TestTextLayoutMetricFailureDoesNotPaint(t *testing.T) {
	base := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	before := append([]byte(nil), base.Pix...)
	failure := errors.New("shape unavailable")
	layout := &softwareTextLayout{
		metricError: failure,
		paint: func(*image.NRGBA, canvas.TextPaintContext) error {
			t.Fatal("paint called after invalid metrics")
			return nil
		},
	}
	text := &canvas.Text{Text: "no fallback", Color: color.Black, Layout: layout}
	drawText(textLayoutCanvas{scale: 1}, text, fyne.Position{}, base, base.Bounds())
	if !bytes.Equal(before, base.Pix) || !errors.Is(text.TextLayoutError(), failure) {
		t.Fatalf("metric failure painted or was hidden: %v", text.TextLayoutError())
	}
}

func TestTextLayoutDecorationsUseSuppliedBaselineAndAlignedAdvance(t *testing.T) {
	base := image.NewNRGBA(image.Rect(0, 0, 40, 30))
	want := image.NewNRGBA(base.Bounds())
	ctx := textLayoutCanvas{scale: 2}
	layout := &softwareTextLayout{
		size: fyne.NewSize(8, 4), baseline: 3,
		paint: func(*image.NRGBA, canvas.TextPaintContext) error {
			return nil
		},
	}
	text := &canvas.Text{Text: "", Color: color.Black, Layout: layout, Alignment: fyne.TextAlignTrailing, TextStyle: fyne.TextStyle{Underline: true, Strikethrough: true}}
	text.Resize(fyne.NewSize(12, 8))
	drawText(ctx, text, fyne.NewPos(2.5, 1), base, base.Bounds())
	// Aligned line box is (6.5,3), with baseline 6. Decorations use its
	// eight-pixel advance even though the fallback Text is empty.
	line := &canvas.Line{Position2: fyne.NewPos(8, 0), StrokeColor: color.Black, StrokeWidth: 1}
	drawLine(ctx, line, fyne.NewPos(6.5, 8), want, want.Bounds())
	drawLine(ctx, line, fyne.NewPos(6.5, 5.25), want, want.Bounds())
	if bytes.Equal(base.Pix, make([]byte, len(base.Pix))) || !bytes.Equal(base.Pix, want.Pix) {
		t.Fatal("decorations did not use supplied baseline, alignment and advance")
	}
}

func TestTextLayoutDecorationsRespectPositiveAncestorClip(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		bounds := image.Rect(0, 0, int(40*scale), int(30*scale))
		full := image.NewNRGBA(bounds)
		clipped := image.NewNRGBA(bounds)
		background := image.NewUniform(color.NRGBA{R: 17, G: 41, B: 67, A: 255})
		draw.Draw(full, bounds, background, image.Point{}, draw.Src)
		draw.Draw(clipped, bounds, background, image.Point{}, draw.Src)
		before := append([]byte(nil), clipped.Pix...)
		layout := &softwareTextLayout{
			size: fyne.NewSize(12, 4), baseline: 3,
			paint: func(*image.NRGBA, canvas.TextPaintContext) error {
				return nil
			},
		}
		text := &canvas.Text{Color: color.White, Layout: layout, TextStyle: fyne.TextStyle{Underline: true, Strikethrough: true}}
		text.Resize(layout.size)
		ctx := textLayoutCanvas{scale: scale}
		pos := fyne.NewPos(1.5, 2)
		drawText(ctx, text, pos, full, bounds)
		clip := image.Rect(int(8*scale), int(4*scale), int(16*scale), int(9*scale))
		drawText(ctx, text, pos, clipped, clip)
		changed := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				offset := clipped.PixOffset(x, y)
				if image.Pt(x, y).In(clip) {
					if clipped.NRGBAAt(x, y) != full.NRGBAAt(x, y) {
						t.Fatalf("scale %v decoration shifted by clip at %d,%d", scale, x, y)
					}
					if !bytes.Equal(clipped.Pix[offset:offset+4], before[offset:offset+4]) {
						changed++
					}
				} else if !bytes.Equal(clipped.Pix[offset:offset+4], before[offset:offset+4]) {
					t.Fatalf("scale %v decoration painted outside clip at %d,%d", scale, x, y)
				}
			}
		}
		if changed == 0 || text.TextLayoutError() != nil {
			t.Fatalf("clipped decorations were empty or failed: %v", text.TextLayoutError())
		}
	}
}

func TestTextLayoutHugeDecorationRejectedBeforeProviderPaint(t *testing.T) {
	base := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	before := append([]byte(nil), base.Pix...)
	calls := 0
	layout := &softwareTextLayout{
		size: fyne.NewSize(math.MaxFloat32, 4), baseline: 3,
		paint: func(*image.NRGBA, canvas.TextPaintContext) error {
			calls++
			return nil
		},
	}
	text := &canvas.Text{Color: color.Black, Layout: layout, TextStyle: fyne.TextStyle{Underline: true}}
	ctx := textLayoutCanvas{scale: 1}
	drawText(ctx, text, fyne.Position{}, base, base.Bounds())
	if calls != 0 || !bytes.Equal(before, base.Pix) || !errors.Is(text.TextLayoutError(), canvas.ErrTextLayoutUnsupportedDecorations) {
		t.Fatalf("unsafe decoration reached provider or was not reported: calls=%d, error=%v", calls, text.TextLayoutError())
	}
	text.TextStyle.Underline = false
	drawText(ctx, text, fyne.Position{}, base, base.Bounds())
	if calls != 1 || text.TextLayoutError() != nil {
		t.Fatalf("undecorated finite advance was incorrectly restricted: calls=%d, error=%v", calls, text.TextLayoutError())
	}
}
