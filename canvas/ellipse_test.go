package canvas_test

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"github.com/stretchr/testify/assert"
)

func TestEllipse_MinSize(t *testing.T) {
	ellipse := canvas.NewEllipse(color.Black)
	min := ellipse.MinSize()

	assert.Positive(t, min.Width)
	assert.Positive(t, min.Height)
}

func TestEllipse_FillColor(t *testing.T) {
	c := color.White
	ellipse := canvas.NewEllipse(c)

	assert.Equal(t, c, ellipse.FillColor)
}

func TestEllipse_Resize(t *testing.T) {
	targetWidth := float32(50)
	targetHeight := float32(80)
	ellipse := canvas.NewEllipse(color.White)
	start := ellipse.Size()
	assert.True(t, start.Height == 0)
	assert.True(t, start.Width == 0)

	ellipse.Resize(fyne.NewSize(targetWidth, targetHeight))
	target := ellipse.Size()
	assert.True(t, target.Height == targetHeight)
	assert.True(t, target.Width == targetWidth)
}

func TestEllipse_Move(t *testing.T) {
	ellipse := canvas.NewEllipse(color.White)
	ellipse.Resize(fyne.NewSize(80, 50))

	start := fyne.Position{X: 0, Y: 0}
	assert.Equal(t, ellipse.Position(), start)

	target := fyne.Position{X: 10, Y: 75}
	ellipse.Move(target)
	assert.Equal(t, ellipse.Position(), target)
}
