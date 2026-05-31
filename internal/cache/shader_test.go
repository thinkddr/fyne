package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
