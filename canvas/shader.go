package canvas

import (
	"image"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/cache"
)

// shaderMaxFrameDelta caps the time advanced per animation frame so that an
// unusually long frame (such as the first one after the animation resumes)
// does not make the shader jump forward.
const shaderMaxFrameDelta = 100 * time.Millisecond

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
//	uniform float time;        // elapsed animation time in seconds (see Start)
//
// and should compute its colour from gl_FragCoord, as the built in shapes do.
// Any images set in Textures are additionally exposed as "uniform sampler2D"
// values, and any values in Uniforms as "uniform float", named by their map key.
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

	// Textures supplies named images to the shader. Each entry is uploaded to
	// the GPU and exposed to the fragment shader as a "uniform sampler2D <name>".
	// Images are uploaded once and reused; replacing an entry with a different
	// image updates the GPU copy on the next paint, so static textures cost
	// nothing per frame.
	Textures map[string]image.Image

	// Uniforms supplies named scalar values to the shader, each exposed to the
	// fragment shader as a "uniform float <name>". They are applied every paint,
	// so an application can drive a shader's parameters - for example animating a
	// transition - by updating an entry and calling Refresh.
	Uniforms map[string]float32

	anim     *fyne.Animation // drives continuous repaints while animating
	running  bool
	elapsed  time.Duration // animation time accumulated while running
	lastTick time.Time     // wall-clock time of the previous animation tick
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

// Start begins animating this shader. While running the shader is redrawn every
// frame and its "time" uniform advances, allowing the fragment shader to produce
// motion. Calling Start on an already running shader has no effect.
//
// Since: 2.8
func (s *Shader) Start() {
	if s.running {
		return
	}

	s.running = true
	s.lastTick = time.Time{} // the next tick re-establishes the clock, so the pause is not counted

	s.anim = fyne.NewAnimation(time.Second, func(float32) {
		s.advance(time.Now())
		s.Refresh()
	})
	s.anim.Curve = fyne.AnimationLinear
	s.anim.RepeatCount = fyne.AnimationRepeatForever
	s.anim.Start()
}

// Stop ends the animation started by Start. The "time" uniform freezes at its
// current value, so the shader keeps its last rendered state until restarted -
// a repaint triggered for any other reason will not advance it.
// Calling Stop on a shader that is not animating has no effect.
//
// Since: 2.8
func (s *Shader) Stop() {
	if !s.running {
		return
	}

	s.running = false

	if s.anim != nil {
		s.anim.Stop()
		s.anim = nil
	}
}

// advance accumulates animation time for the current frame and publishes it for
// the painter. It is only called from the animation tick, so the shader's time
// progresses solely while it is running, never on an unrelated repaint.
func (s *Shader) advance(now time.Time) {
	if !s.lastTick.IsZero() {
		delta := now.Sub(s.lastTick)
		if delta > shaderMaxFrameDelta {
			delta = shaderMaxFrameDelta
		}
		s.elapsed += delta
	}
	s.lastTick = now
	cache.SetShaderTime(s.Name, float32(s.elapsed.Seconds()))
}
