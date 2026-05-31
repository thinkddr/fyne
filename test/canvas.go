package test

import (
	"image"
	"image/draw"
	"sync"

	"fyne.io/fyne/v2"
	fynedriver "fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/internal"
	intapp "fyne.io/fyne/v2/internal/app"
	"fyne.io/fyne/v2/internal/cache"
	"fyne.io/fyne/v2/internal/scale"
	"fyne.io/fyne/v2/theme"
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
	return wrapCanvas(newSoftwareCanvas(nil, false))
}

// NewCanvasWithPainter allows creation of an in-memory canvas with a specific painter.
// The painter will be used to render in the Capture() call.
func NewCanvasWithPainter(painter fynedriver.Painter) WindowlessCanvas {
	return wrapCanvas(newSoftwareCanvas(painter, false))
}

// NewTransparentCanvasWithPainter allows creation of an in-memory canvas with a specific painter without a background color.
// The painter will be used to render in the Capture() call.
//
// Since: 2.2
func NewTransparentCanvasWithPainter(painter fynedriver.Painter) WindowlessCanvas {
	return wrapCanvas(newSoftwareCanvas(painter, true))
}

func newSoftwareCanvas(painter fynedriver.Painter, transparent bool) *softwareCanvas {
	c := &softwareCanvas{
		focusMgr:    intapp.NewFocusManager(nil),
		padded:      true,
		painter:     painter,
		scale:       1.0,
		size:        fyne.NewSize(100, 100),
		transparent: transparent,
	}
	c.overlays.Canvas = c
	return c
}

func wrapCanvas(c WindowlessCanvas) *canvas {
	return &canvas{WindowlessCanvas: c}
}

type canvas struct {
	WindowlessCanvas
	hovered desktop.Hoverable
}
type softwareCanvas struct {
	size    fyne.Size
	resized bool
	scale   float32

	content     fyne.CanvasObject
	overlays    internal.OverlayStack
	focusMgr    *intapp.FocusManager
	padded      bool
	transparent bool

	onTypedRune func(rune)
	onTypedKey  func(*fyne.KeyEvent)

	fyne.ShortcutHandler
	painter      fynedriver.Painter
	propertyLock sync.RWMutex
}

func (c *softwareCanvas) Capture() image.Image {
	cache.Clean(true)
	size := c.Size()
	bounds := image.Rect(0, 0, scale.ToScreenCoordinate(c, size.Width), scale.ToScreenCoordinate(c, size.Height))
	img := image.NewNRGBA(bounds)
	if !c.transparent {
		draw.Draw(img, bounds, image.NewUniform(theme.Color(theme.ColorNameBackground)), image.Point{}, draw.Src)
	}

	if c.painter != nil {
		draw.Draw(img, bounds, c.painter.Paint(c), image.Point{}, draw.Over)
	}

	return img
}

func (c *softwareCanvas) Content() fyne.CanvasObject {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.content
}

func (c *softwareCanvas) Focus(obj fyne.Focusable) {
	c.focusManager().Focus(obj)
}

func (c *softwareCanvas) FocusNext() {
	c.focusManager().FocusNext()
}

func (c *softwareCanvas) FocusPrevious() {
	c.focusManager().FocusPrevious()
}

func (c *softwareCanvas) Focused() fyne.Focusable {
	return c.focusManager().Focused()
}

func (c *softwareCanvas) InteractiveArea() (fyne.Position, fyne.Size) {
	return fyne.NewPos(0, 0), c.Size()
}

func (c *softwareCanvas) OnTypedKey() func(*fyne.KeyEvent) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.onTypedKey
}

func (c *softwareCanvas) OnTypedRune() func(rune) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.onTypedRune
}

func (c *softwareCanvas) Overlays() fyne.OverlayStack {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	return &c.overlays
}

func (c *softwareCanvas) Padded() bool {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.padded
}

func (c *softwareCanvas) PixelCoordinateForPosition(pos fyne.Position) (int, int) {
	return int(pos.X * c.scale), int(pos.Y * c.scale)
}

func (c *softwareCanvas) Refresh(fyne.CanvasObject) {
}

func (c *softwareCanvas) Resize(size fyne.Size) {
	c.propertyLock.Lock()
	c.resized = true
	c.propertyLock.Unlock()

	c.doResize(size)
}

func (c *softwareCanvas) doResize(size fyne.Size) {
	c.propertyLock.Lock()
	content := c.content
	overlays := c.overlays
	padded := c.padded
	c.size = size
	c.propertyLock.Unlock()

	if content == nil {
		return
	}

	// Ensure testcanvas mimics real canvas.Resize behavior
	fullPos, fullSize := c.InteractiveArea()
	for _, overlay := range overlays.List() {
		overlay.Move(fullPos)
		overlay.Resize(fullSize)
	}

	if padded {
		padding := theme.Padding()
		content.Resize(size.Subtract(fyne.NewSquareSize(padding * 2)))
		content.Move(fyne.NewSquareOffsetPos(padding))
	} else {
		content.Resize(size)
		content.Move(fyne.NewPos(0, 0))
	}
}

func (c *softwareCanvas) Scale() float32 {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.scale
}

func (c *softwareCanvas) SetContent(content fyne.CanvasObject) {
	c.propertyLock.Lock()
	c.content = content
	c.focusMgr = intapp.NewFocusManager(c.content)
	resized := c.resized
	c.propertyLock.Unlock()

	if content == nil {
		return
	}

	minSize := content.MinSize()
	if c.padded {
		minSize = minSize.Add(fyne.NewSquareSize(theme.Padding() * 2))
	}

	if resized {
		c.doResize(c.Size().Max(minSize))
	} else {
		c.doResize(minSize)
	}
}

func (c *softwareCanvas) SetOnTypedKey(handler func(*fyne.KeyEvent)) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.onTypedKey = handler
}

func (c *softwareCanvas) SetOnTypedRune(handler func(rune)) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.onTypedRune = handler
}

func (c *softwareCanvas) SetPadded(padded bool) {
	c.propertyLock.Lock()
	c.padded = padded
	c.propertyLock.Unlock()

	c.doResize(c.Size())
}

func (c *softwareCanvas) SetScale(scale float32) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.scale = scale
}

func (c *softwareCanvas) Size() fyne.Size {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.size
}

func (c *softwareCanvas) Unfocus() {
	c.focusManager().Focus(nil)
}

func (c *softwareCanvas) focusManager() *intapp.FocusManager {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if focusMgr := c.overlays.TopFocusManager(); focusMgr != nil {
		return focusMgr
	}
	return c.focusMgr
}
