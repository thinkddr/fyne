package canvas_test

import (
	"errors"
	"image"
	"image/color"
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
)

type testTextLayout struct {
	size     fyne.Size
	baseline float32
	err      error
}

func (l *testTextLayout) Metrics() (fyne.Size, float32, error) {
	return l.size, l.baseline, l.err
}

func (*testTextLayout) PaintSoftware(*image.NRGBA, canvas.TextPaintContext) error {
	panic("metric queries must not paint")
}

func TestTextLayoutAuthoritativeMetrics(t *testing.T) {
	layout := &testTextLayout{size: fyne.NewSize(27.25, 3.5), baseline: 11.75}
	text := &canvas.Text{Text: "Default font metrics must not be used", TextSize: 100, Layout: layout}
	if got := text.MinSize(); got != layout.size {
		t.Fatalf("MinSize = %v, want supplied %v", got, layout.size)
	}
	size, baseline, err := text.LayoutMetrics()
	if err != nil || size != layout.size || baseline != 11.75 {
		t.Fatalf("LayoutMetrics = %v, %v, %v", size, baseline, err)
	}
	layout.baseline = -2
	if _, baseline, err = text.LayoutMetrics(); err != nil || baseline != -2 {
		t.Fatalf("finite baseline outside the line box rejected: %v, %v", baseline, err)
	}
}

func TestTextLayoutInvalidMetricsDoNotFallback(t *testing.T) {
	for _, tc := range []struct {
		name     string
		size     fyne.Size
		baseline float32
	}{
		{"negative width", fyne.NewSize(-1, 10), 8},
		{"negative height", fyne.NewSize(10, -1), 8},
		{"infinite width", fyne.NewSize(float32(math.Inf(1)), 10), 8},
		{"nan height", fyne.NewSize(10, float32(math.NaN())), 8},
		{"nan baseline", fyne.NewSize(10, 10), float32(math.NaN())},
		{"infinite baseline", fyne.NewSize(10, 10), float32(math.Inf(-1))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text := &canvas.Text{Text: "visible fallback is forbidden", Layout: &testTextLayout{size: tc.size, baseline: tc.baseline}}
			if got := text.MinSize(); got != (fyne.Size{}) {
				t.Fatalf("invalid layout returned nonzero size: %v", got)
			}
			if !errors.Is(text.TextLayoutError(), canvas.ErrTextLayoutInvalidMetrics) {
				t.Fatalf("invalid metric error unavailable: %v", text.TextLayoutError())
			}
		})
	}
}

func TestTextLayoutMetricAndPaintErrorsRemainDistinct(t *testing.T) {
	metricFailure := errors.New("font engine is closed")
	paintFailure := errors.New("destination format is unsupported")
	layout := &testTextLayout{size: fyne.NewSize(20, 10), baseline: 8, err: metricFailure}
	text := &canvas.Text{Layout: layout}
	text.SetTextLayoutError(paintFailure)
	if _, _, err := text.LayoutMetrics(); !errors.Is(err, metricFailure) {
		t.Fatalf("metric failure lost: %v", err)
	}
	if err := text.TextLayoutError(); !errors.Is(err, metricFailure) || !errors.Is(err, paintFailure) {
		t.Fatalf("combined error lost a cause: %v", err)
	}
	layout.err = nil
	if got := text.MinSize(); got != layout.size {
		t.Fatalf("valid replacement metrics: %v", got)
	}
	if err := text.TextLayoutError(); !errors.Is(err, paintFailure) || errors.Is(err, metricFailure) {
		t.Fatalf("metric success cleared paint failure or retained metric failure: %v", err)
	}
	text.SetTextLayoutError(nil)
	if err := text.TextLayoutError(); err != nil {
		t.Fatalf("explicit paint reset failed: %v", err)
	}
}

func TestTextLayoutNilKeepsDriverMetrics(t *testing.T) {
	test.NewTempApp(t)
	text := canvas.NewText("Unmodified default shaping", color.Black)
	wantSize, wantBaseline := fyne.CurrentApp().Driver().RenderedTextSize(text.Text, text.TextSize, text.TextStyle, text.FontSource)
	if got := text.MinSize(); got != wantSize {
		t.Fatalf("default MinSize changed: %v, want %v", got, wantSize)
	}
	size, baseline, err := text.LayoutMetrics()
	if err != nil || size != wantSize || baseline != wantBaseline {
		t.Fatalf("default LayoutMetrics = %v, %v, %v", size, baseline, err)
	}
}
