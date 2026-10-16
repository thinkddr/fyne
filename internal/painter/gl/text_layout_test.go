package gl

import (
	"errors"
	"image"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type unsupportedTextLayout struct{}

func (unsupportedTextLayout) Metrics() (fyne.Size, float32, error) {
	return fyne.NewSize(7, 10), 8, nil
}

func (unsupportedTextLayout) PaintSoftware(*image.NRGBA, canvas.TextPaintContext) error {
	panic("GL must not call a software provider")
}

func TestTextLayoutGLRejectsWithoutFallback(t *testing.T) {
	for _, value := range []string{"", " ", "not default glyphs"} {
		text := &canvas.Text{Text: value, Layout: unsupportedTextLayout{}}
		// An uninitialized painter has no GL context: entering the default path
		// or allocating a texture would panic instead of reporting this error.
		p := &painter{}
		p.drawText(text, fyne.Position{}, fyne.NewSize(30, 20), nil)
		if !errors.Is(text.TextLayoutError(), canvas.ErrTextLayoutUnsupportedGL) {
			t.Fatalf("GL did not reject supplied layout for %q: %v", value, text.TextLayoutError())
		}
		if text.MinSize() != fyne.NewSize(7, 10) || !errors.Is(text.TextLayoutError(), canvas.ErrTextLayoutUnsupportedGL) {
			t.Fatalf("metric query hid unsupported GL paint: %v", text.TextLayoutError())
		}
	}
}
