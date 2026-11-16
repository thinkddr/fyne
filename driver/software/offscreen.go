package software

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/cache"
	"fyne.io/fyne/v2/internal/driver"
	"fyne.io/fyne/v2/internal/painter/software"
)

// Region is the absolute bounds of an object reported by RenderOffscreen, in the
// canvas units of the render (not pixels).
//
// Since: 2.8.2-sytue.6
type Region struct {
	Object   fyne.CanvasObject
	Position fyne.Position
	Size     fyne.Size
}

// Offscreen is one render made by RenderOffscreen.
//
// Since: 2.8.2-sytue.6
type Offscreen struct {
	// Image has Size × Scale pixels and a transparent background.
	Image *image.NRGBA
	// Size is the size the content was laid out at.
	Size fyne.Size
	// Scale is the pixel density used to paint.
	Scale float32
	// Regions are the visible objects for which the mark function returned true, in
	// painting order, with their absolute position inside the render.
	Regions []Region
}

// RenderOffscreen lays obj out at size and paints it once at scale onto a transparent
// image, off any window. It is meant for surfaces that are drawn by another process,
// such as home screen widgets, which can only show a picture and a list of tap areas.
//
// Unlike Render and RenderCanvas it does not change the application theme, does not
// paint the theme background and does not run the global cache cleaner, so it is safe
// to call while the app shows windows, from the goroutine that runs the app (use
// fyne.DoAndWait from elsewhere). Without an application it can be called from any
// single goroutine. The renderers it creates are destroyed before returning, so obj
// must not be shown in a window at the same time.
//
// mark may be nil. Otherwise every visible object in the tree is offered to it, and the
// ones it accepts are reported in Regions; their children are still walked.
//
// Since: 2.8.2-sytue.6
func RenderOffscreen(obj fyne.CanvasObject, size fyne.Size, scale float32, mark func(fyne.CanvasObject) bool) Offscreen {
	if scale <= 0 {
		scale = 1
	}
	out := Offscreen{Size: size, Scale: scale}
	if obj == nil || size.Width <= 0 || size.Height <= 0 {
		out.Image = image.NewNRGBA(image.Rect(0, 0, 0, 0))
		return out
	}

	c := newCanvas(software.NewPainter(), true)
	c.SetPadded(false)
	c.SetScale(scale)
	c.SetContent(obj)
	c.Resize(size)

	painted := c.(*canvas).painter.Paint(c)
	out.Image = toNRGBA(painted)

	if mark != nil {
		driver.WalkVisibleObjectTree(obj, func(o fyne.CanvasObject, pos, _ fyne.Position, _ fyne.Size) bool {
			if mark(o) {
				out.Regions = append(out.Regions, Region{Object: o, Position: pos, Size: o.Size()})
			}
			return false
		}, nil)
	}

	// The render is done: nothing of it must stay in the caches the windows use.
	// After the children: walking a widget's children asks for its renderer, which would
	// create a new one if it were already destroyed.
	driver.WalkCompleteObjectTree(obj, func(fyne.CanvasObject, fyne.Position, fyne.Position, fyne.Size) bool {
		return false
	}, func(o fyne.CanvasObject, _ fyne.Position, _ fyne.CanvasObject) {
		if w, ok := o.(fyne.Widget); ok {
			cache.DestroyRenderer(w)
		}
	})
	cache.CleanCanvas(c)
	return out
}

func toNRGBA(img image.Image) *image.NRGBA {
	if n, ok := img.(*image.NRGBA); ok {
		return n
	}
	b := img.Bounds()
	n := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			n.Set(x, y, img.At(x, y))
		}
	}
	return n
}
