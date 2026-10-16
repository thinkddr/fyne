package software

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/internal/painter"
	"fyne.io/fyne/v2/theme"
)

func drawTextLayout(c fyne.Canvas, text *canvas.Text, pos fyne.Position, base *image.NRGBA, clip image.Rectangle) {
	bounds, baseline, err := text.LayoutMetrics()
	if err != nil {
		fyne.LogError("Failed to measure canvas.Text layout", err)
		return
	}
	origin := pos
	size := text.Size()
	switch text.Alignment {
	case fyne.TextAlignTrailing:
		origin.X += size.Width - bounds.Width
	case fyne.TextAlignCenter:
		origin.X += (size.Width - bounds.Width) / 2
	}
	if size.Height > bounds.Height {
		origin.Y += (size.Height - bounds.Height) / 2
	}
	scale := c.Scale()
	underlineY := origin.Y + baseline + painter.UnderlineOffsetFromBaseline
	strikeY := origin.Y + baseline*painter.StrikethroughToBaselineFactor
	if !finiteLayoutCoordinate(origin.X) || !finiteLayoutCoordinate(origin.Y) || !finiteLayoutCoordinate(scale) || scale <= 0 ||
		(text.TextStyle.Underline && !finiteLayoutCoordinate(underlineY)) ||
		(text.TextStyle.Strikethrough && !finiteLayoutCoordinate(strikeY)) {
		err = errors.New("canvas.Text layout paint has invalid coordinates or scale")
		text.SetTextLayoutError(err)
		fyne.LogError("Failed to paint canvas.Text layout", err)
		return
	}
	clip = clip.Intersect(base.Bounds())
	if clip.Empty() {
		return
	}
	ink := text.Color
	if ink == nil {
		ink = theme.Color(theme.ColorNameForeground)
	}
	var decorations []textLayoutDecoration
	for _, decoration := range []struct {
		enabled bool
		y       float32
	}{{text.TextStyle.Underline, underlineY}, {text.TextStyle.Strikethrough, strikeY}} {
		if !decoration.enabled {
			continue
		}
		line, err := newTextLayoutDecoration(bounds.Width, fyne.NewPos(origin.X, decoration.y), scale, ink)
		if err != nil {
			text.SetTextLayoutError(err)
			fyne.LogError("Failed to paint canvas.Text layout", err)
			return
		}
		decorations = append(decorations, line)
	}
	context := canvas.TextPaintContext{Origin: origin, Scale: scale, Clip: clip, Color: ink}
	destination := base.SubImage(clip).(*image.NRGBA)
	if err = text.Layout.PaintSoftware(destination, context); err != nil {
		text.SetTextLayoutError(err)
		fyne.LogError("Failed to paint canvas.Text layout", err)
		return
	}
	for _, decoration := range decorations {
		decoration.paint(scale, base, clip)
	}
	text.SetTextLayoutError(nil)
}

type textLayoutDecoration struct {
	line   *canvas.Line
	pad    float32
	bounds image.Rectangle
}

func newTextLayoutDecoration(width float32, pos fyne.Position, scale float32, ink color.Color) (textLayoutDecoration, error) {
	line := &canvas.Line{StrokeColor: ink, StrokeWidth: 1, Position2: fyne.NewPos(width, 0)}
	pad := painter.VectorPad(line)
	w := math.Round(float64(width+pad*2) * float64(scale))
	h := math.Round(float64(pad*2) * float64(scale))
	x := math.Round(float64(pos.X-pad) * float64(scale))
	y := math.Round(float64(pos.Y-pad) * float64(scale))
	// DrawLine uses integer scanner bounds and 26.6 fixed-point coordinates.
	// Bound its temporary raster to 16M pixels before the provider can write.
	const maxCoordinate = (1 << 31) - 1
	if float64(width) > maxCoordinate || w < 1 || h < 1 || w*h > 16*1024*1024 ||
		x < -maxCoordinate || y < -maxCoordinate || x+w > maxCoordinate || y+h > maxCoordinate {
		return textLayoutDecoration{}, canvas.ErrTextLayoutUnsupportedDecorations
	}
	return textLayoutDecoration{line: line, pad: pad, bounds: image.Rect(int(x), int(y), int(x+w), int(y+h))}, nil
}

func (d textLayoutDecoration) paint(scale float32, base *image.NRGBA, clip image.Rectangle) {
	raw := painter.DrawLine(d.line, d.pad, func(value float32) float32 {
		return float32(math.Round(float64(value) * float64(scale)))
	})
	bounds := clip.Intersect(d.bounds)
	source := raw.Bounds().Min.Add(bounds.Min.Sub(d.bounds.Min))
	draw.Draw(base, bounds, raw, source, draw.Over)
}

func finiteLayoutCoordinate(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}
