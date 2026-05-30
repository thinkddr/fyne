package canvas

import (
	"fyne.io/fyne/v2"
)

// Declare conformity with CanvasObject interface
var _ fyne.CanvasObject = (*Shader)(nil)

// Shader describes a canvas object that is drawn using a custom GLSL fragment
// shader. The shader is rendered inside the object's bounds, respecting the
// position and size set on it just like any other canvas primitive.
//
// The supplied fragment shader must follow the same conventions as Fyne's
// internal vector shaders. It is provided with the following uniforms:
//
//	uniform vec2 frame_size;   // the size of the output frame, in pixels
//	uniform vec4 rect_coords;  // this object's bounds (x1, y1, x2, y2), in pixels
//	uniform float edge_softness;
//
// and should compute its colour from gl_FragCoord, as the built in shapes do.
//
// Two source variants are held so that the object renders on both desktop
// OpenGL (core profile) and OpenGL ES / mobile / web targets.
//
// Since: 2.8
type Shader struct {
	baseObject

	// Name identifies this shader. Shaders that share a Name may share a
	// compiled program, so it should be unique per distinct shader source.
	Name string

	// Source is the GLSL fragment shader used on desktop OpenGL (core profile).
	Source []byte

	// SourceES is the GLSL fragment shader used on OpenGL ES, mobile and web.
	SourceES []byte
}

// NewShader returns a new Shader instance using the specified fragment shader
// sources. The name should uniquely identify the shader, source is used on
// desktop (OpenGL core profile) and sourceES is used on mobile and web targets.
//
// Since: 2.8
func NewShader(name string, source, sourceES []byte) *Shader {
	return &Shader{Name: name, Source: source, SourceES: sourceES}
}

// Hide will set this shader to not be visible.
func (s *Shader) Hide() {
	s.baseObject.Hide()

	repaint(s)
}

// Move the shader object to a new position, relative to its parent / canvas.
func (s *Shader) Move(pos fyne.Position) {
	if s.Position() == pos {
		return
	}

	s.baseObject.Move(pos)

	repaint(s)
}

// Refresh causes this shader to be redrawn with its current state.
func (s *Shader) Refresh() {
	Refresh(s)
}

// Resize on a shader updates the new size of this object.
func (s *Shader) Resize(size fyne.Size) {
	if size == s.Size() {
		return
	}

	s.baseObject.Resize(size)

	Refresh(s)
}
