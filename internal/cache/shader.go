package cache

import (
	"time"

	"fyne.io/fyne/v2/internal/async"
)

// shaderTimes holds the current animation time, in seconds, for each named
// shader. It is written by canvas.Shader while it animates and read by the GL
// painter to feed the shader's "time" uniform. Keying by name (rather than the
// object) means the entry is bounded by the number of distinct shaders and does
// not retain canvas objects, matching how compiled shader programs are cached.
var shaderTimes async.Map[string, *shaderTimeInfo]

type shaderTimeInfo struct {
	expiringCache
	seconds float32
}

// SetShaderTime records the current animation time, in seconds, for the named shader.
func SetShaderTime(name string, seconds float32) {
	info := &shaderTimeInfo{seconds: seconds}
	info.setAlive()
	shaderTimes.Store(name, info)
}

// ShaderTime returns the animation time, in seconds, last recorded for the named
// shader, or zero if it has never animated.
func ShaderTime(name string) float32 {
	info, ok := shaderTimes.Load(name)
	if info == nil || !ok {
		return 0
	}
	info.setAlive()
	return info.seconds
}

// destroyExpiredShaders removes shader time cache entries that have not been
// used recently.
func destroyExpiredShaders(now time.Time) {
	shaderTimes.Range(func(name string, info *shaderTimeInfo) bool {
		if info.isExpired(now) {
			shaderTimes.Delete(name)
		}
		return true
	})
}
