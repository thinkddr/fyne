package canvas

import (
	"image/color"

	"fyne.io/fyne/v2"
)

// Declare conformity with CanvasObject interface
var _ fyne.CanvasObject = (*ArbitraryPolygon)(nil)

// ArbitraryPolygon describes a colored arbitrary polygon primitive in a Fyne canvas.
// The polygon is defined by a list of vertex positions in clockwise order, specified in absolute coordinates
// relative to the object (top-left is (0,0), bottom-right is (width,height)).
// Each corner can have an individually specified rounding radius.
//
// Since: 2.8
type ArbitraryPolygon struct {
	baseObject

	Points           []fyne.Position // Vertices in coordinates relative to the object. If NormalizedPoints is true, these are (0.0 to 1.0), otherwise absolute.
	NormalizedPoints bool            // True if Points are specified in normalized coordinates (0.0 to 1.0) relative to the object's size.
	CornerRadii      []float32       // Per-corner rounding radius, must match len(Points); missing entries default to 0
	FillColor        color.Color     // The polygon fill color
	StrokeColor      color.Color     // The polygon stroke color
	StrokeWidth      float32         // The stroke width of the polygon
}

// Hide will set this arbitrary polygon to not be visible
func (r *ArbitraryPolygon) Hide() {
	r.baseObject.Hide()

	repaint(r)
}

// Move the arbitrary polygon to a new position, relative to its parent / canvas
func (r *ArbitraryPolygon) Move(pos fyne.Position) {
	if r.Position() == pos {
		return
	}

	r.baseObject.Move(pos)

	repaint(r)
}

// Refresh causes this arbitrary polygon to be redrawn with its configured state.
func (r *ArbitraryPolygon) Refresh() {
	Refresh(r)
}

// Resize on an arbitrary polygon updates the new size of this object.
func (r *ArbitraryPolygon) Resize(s fyne.Size) {
	if s == r.Size() {
		return
	}

	r.baseObject.Resize(s)

	Refresh(r)
}

// NewArbitraryPolygon returns a new ArbitraryPolygon instance
func NewArbitraryPolygon(points []fyne.Position, fill color.Color) *ArbitraryPolygon {
	return &ArbitraryPolygon{
		Points:    points,
		FillColor: fill,
	}
}
