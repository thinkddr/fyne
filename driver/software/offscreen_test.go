package software_test

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/software"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func TestRenderOffscreen_TransparentAtScale(t *testing.T) {
	red := canvas.NewRectangle(color.NRGBA{R: 0xff, A: 0xff})
	red.SetMinSize(fyne.NewSize(10, 10))
	obj := container.NewWithoutLayout(red)
	red.Move(fyne.NewPos(5, 5))
	red.Resize(fyne.NewSize(10, 10))

	out := software.RenderOffscreen(obj, fyne.NewSize(20, 20), 2, nil)
	if got := out.Image.Bounds().Size(); got.X != 40 || got.Y != 40 {
		t.Fatalf("image size %v, want 40×40 at scale 2", got)
	}
	if a := out.Image.NRGBAAt(2, 2).A; a != 0 {
		t.Fatalf("the background is painted (alpha %d): it must stay transparent", a)
	}
	if px := out.Image.NRGBAAt(20, 20); px.R != 0xff || px.A != 0xff {
		t.Fatalf("the rectangle is not painted at its scaled position: %v", px)
	}
}

func TestRenderOffscreen_ReportsMarkedRegionsAbsolutely(t *testing.T) {
	target := canvas.NewRectangle(color.Black)
	other := canvas.NewRectangle(color.White)
	inner := container.NewWithoutLayout(target)
	target.Move(fyne.NewPos(3, 4))
	target.Resize(fyne.NewSize(6, 7))
	outer := container.NewWithoutLayout(other, inner)
	inner.Move(fyne.NewPos(10, 20))
	inner.Resize(fyne.NewSize(30, 30))

	out := software.RenderOffscreen(outer, fyne.NewSize(50, 60), 1, func(o fyne.CanvasObject) bool { return o == target })
	if len(out.Regions) != 1 {
		t.Fatalf("regions %+v, want only the marked one", out.Regions)
	}
	r := out.Regions[0]
	if r.Position != fyne.NewPos(13, 24) || r.Size != fyne.NewSize(6, 7) {
		t.Fatalf("region at %v size %v, want (13,24) 6×7", r.Position, r.Size)
	}
}

func TestRenderOffscreen_LeavesTheAppThemeAlone(t *testing.T) {
	a := test.NewTempApp(t)
	before := a.Settings().Theme()
	label := widget.NewLabel("Hola")
	software.RenderOffscreen(label, fyne.NewSize(80, 30), 1, nil)
	if a.Settings().Theme() != before {
		t.Fatal("RenderOffscreen changed the application theme")
	}
	_ = theme.Padding()
}

func TestRenderOffscreen_EmptyInputs(t *testing.T) {
	if out := software.RenderOffscreen(nil, fyne.NewSize(10, 10), 1, nil); out.Image == nil || out.Image.Bounds().Dx() != 0 {
		t.Fatal("nil content must give an empty image")
	}
	if out := software.RenderOffscreen(canvas.NewRectangle(color.Black), fyne.NewSize(0, 10), 1, nil); out.Image.Bounds().Dx() != 0 {
		t.Fatal("an empty size must give an empty image")
	}
}
