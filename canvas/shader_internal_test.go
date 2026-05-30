package canvas

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/internal/cache"

	"github.com/stretchr/testify/assert"
)

func TestShader_advance(t *testing.T) {
	s := &Shader{Name: "test-advance"}
	base := time.Unix(100, 0)

	// the first tick establishes the clock and publishes zero elapsed
	s.advance(base)
	assert.Equal(t, float32(0), cache.ShaderTime(s.Name))

	// a normal tick advances the published time by its own duration
	s.advance(base.Add(16 * time.Millisecond))
	assert.InDelta(t, 0.016, cache.ShaderTime(s.Name), 1e-4)

	// an unusually long tick is capped so the shader does not jump forward
	s.advance(base.Add(16*time.Millisecond + 5*time.Second))
	assert.InDelta(t, 0.016+shaderMaxFrameDelta.Seconds(), float64(cache.ShaderTime(s.Name)), 1e-4)

	// Start resets lastTick; the first tick after resuming must not count the
	// time that elapsed while the shader was stopped
	resumeFrom := cache.ShaderTime(s.Name)
	s.lastTick = time.Time{}
	s.advance(base.Add(time.Hour))
	assert.Equal(t, resumeFrom, cache.ShaderTime(s.Name))
}
