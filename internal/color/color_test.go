package color_test

import (
	imagecolor "image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2/internal/color"
)

func Test_ToNRGBA_unmultiplyAlpha(t *testing.T) {
	c := imagecolor.RGBA{R: 100, G: 100, B: 100, A: 100}
	r, g, b, a := color.ToNRGBA(c)

	assert.Equal(t, 255, r)
	assert.Equal(t, 255, g)
	assert.Equal(t, 255, b)
	assert.Equal(t, 100, a)

	c = imagecolor.RGBA{R: 100, G: 100, B: 100, A: 255}
	r, g, b, a = color.ToNRGBA(c)

	assert.Equal(t, 100, r)
	assert.Equal(t, 100, g)
	assert.Equal(t, 100, b)
	assert.Equal(t, 255, a)
}
