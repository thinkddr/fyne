package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDestroyExpiredShaders(t *testing.T) {
	tm := &timeMock{}
	tm.setTime(0, 0)
	defer testClearAll()

	SetShaderTime("shader-stale", 1.5)

	// not yet expired
	destroyExpiredShaders(tm.createTime(0, 30))
	assert.Equal(t, float32(1.5), ShaderTime("shader-stale"))

	// reading kept it alive, so it survives just past the original window
	tm.setTime(1, 0)
	destroyExpiredShaders(tm.createTime(1, 0))
	assert.Equal(t, float32(1.5), ShaderTime("shader-stale"))

	// no more access, it expires and is removed
	destroyExpiredShaders(tm.createTime(5, 0))
	_, ok := shaderTimes.Load("shader-stale")
	assert.False(t, ok, "stale shader was not cleaned up")
	assert.Equal(t, float32(0), ShaderTime("shader-stale"))
}

func TestShaderTime_unknown(t *testing.T) {
	assert.Equal(t, float32(0), ShaderTime("shader-never-set"))
}

func TestSetShaderTime(t *testing.T) {
	SetShaderTime("shader-time", 1.5)
	assert.Equal(t, float32(1.5), ShaderTime("shader-time"))

	SetShaderTime("shader-time", 2.25)
	assert.Equal(t, float32(2.25), ShaderTime("shader-time"), "value did not replace old")

	SetShaderTime("shader-other", 4)
	assert.Equal(t, float32(2.25), ShaderTime("shader-time"), "previous value was lost")
	assert.Equal(t, float32(4), ShaderTime("shader-other"))
}
