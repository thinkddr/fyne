package cache

import "sync"

// shaderTimes holds the current animation time, in seconds, for each named
// shader. It is written by canvas.Shader while it animates and read by the GL
// painter to feed the shader's "time" uniform. Keying by name (rather than the
// object) means the entry is bounded by the number of distinct shaders and does
// not retain canvas objects, matching how compiled shader programs are cached.
var shaderTimes sync.Map // map[string]float32

// SetShaderTime records the current animation time, in seconds, for the named shader.
func SetShaderTime(name string, seconds float32) {
	shaderTimes.Store(name, seconds)
}

// ShaderTime returns the animation time, in seconds, last recorded for the named
// shader, or zero if it has never animated.
func ShaderTime(name string) float32 {
	if v, ok := shaderTimes.Load(name); ok {
		return v.(float32)
	}
	return 0
}
