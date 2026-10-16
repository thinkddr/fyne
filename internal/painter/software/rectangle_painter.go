// SPDX-License-Identifier: BSD-3-Clause

package software

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

func drawSuppliedRectangle(c fyne.Canvas, rectangle *canvas.Rectangle, origin fyne.Position, base *image.NRGBA, clip image.Rectangle) {
	size, scale := rectangle.Size(), c.Scale()
	if !finiteLayoutCoordinate(origin.X) || !finiteLayoutCoordinate(origin.Y) ||
		!finiteLayoutCoordinate(size.Width) || !finiteLayoutCoordinate(size.Height) || size.Width < 0 || size.Height < 0 ||
		!finiteLayoutCoordinate(scale) || scale <= 0 {
		rectangle.SetPaintError(canvas.ErrRectanglePainterInvalidGeometry)
		fyne.LogError("Failed to paint canvas.Rectangle", canvas.ErrRectanglePainterInvalidGeometry)
		return
	}
	clip = clip.Intersect(base.Bounds())
	if clip.Empty() {
		return
	}
	context := canvas.RectanglePaintContext{Origin: origin, Size: size, Scale: scale, Clip: clip}
	err := rectangle.Painter.PaintSoftware(base.SubImage(clip).(*image.NRGBA), context)
	rectangle.SetPaintError(err)
	if err != nil {
		fyne.LogError("Failed to paint canvas.Rectangle", err)
	}
}
