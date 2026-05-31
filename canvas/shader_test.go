package canvas_test

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"

	"github.com/stretchr/testify/assert"
)

func TestNewShader(t *testing.T) {
	src := []byte("core source")
	srcES := []byte("es source")
	shader := canvas.NewShader("test", src, srcES)

	assert.Equal(t, "test", shader.Name)
	assert.Equal(t, src, shader.Source)
	assert.Equal(t, srcES, shader.SourceES)
	assert.True(t, shader.Visible())
}

func TestShader_MinSize(t *testing.T) {
	shader := canvas.NewShader("test", nil, nil)
	min := shader.MinSize()

	assert.Positive(t, min.Width)
	assert.Positive(t, min.Height)
}

func TestShader_Resize(t *testing.T) {
	shader := canvas.NewShader("test", nil, nil)
	size := fyne.NewSize(100, 50)
	shader.Resize(size)

	assert.Equal(t, size, shader.Size())
}

func TestShader_Move(t *testing.T) {
	shader := canvas.NewShader("test", nil, nil)
	pos := fyne.NewPos(10, 20)
	shader.Move(pos)

	assert.Equal(t, pos, shader.Position())
}

func TestShader_Hide(t *testing.T) {
	shader := canvas.NewShader("test", nil, nil)
	assert.True(t, shader.Visible())

	shader.Hide()
	assert.False(t, shader.Visible())

	shader.Show()
	assert.True(t, shader.Visible())
}

func TestShader_Refresh(t *testing.T) {
	test.NewTempApp(t)
	shader := canvas.NewShader("test", nil, nil)

	assert.NotPanics(t, shader.Refresh)
}

func TestShader_StartStop(t *testing.T) {
	test.NewTempApp(t)
	shader := canvas.NewShader("test", nil, nil)

	assert.NotPanics(t, shader.Stop, "Stopping a shader before start should be no-op")

	assert.NotPanics(t, shader.Start)
	assert.NotPanics(t, shader.Start, "Starting a started shader should be a no-op")

	assert.NotPanics(t, shader.Stop)
	assert.NotPanics(t, shader.Stop, "Stopping a stopped shader should be no-op")

	assert.NotPanics(t, shader.Start, "Re-starting a shader should function correctly")
	assert.NotPanics(t, shader.Stop)
}
