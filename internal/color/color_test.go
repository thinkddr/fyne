package color_test

import (
	imagecolor "image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2/internal/color"
)

func Test_ToNRGBA_unmultiplyAlpha(t *testing.T) {
	for name, tt := range map[string]struct {
		color imagecolor.Color
		wantR int
		wantG int
		wantB int
		wantA int
	}{
		"RGBA": {
			color: imagecolor.RGBA{R: 100, G: 100, B: 100, A: 100},
			wantR: 255,
			wantG: 255,
			wantB: 255,
			wantA: 100,
		},
		"RGBA opaque": {
			color: imagecolor.RGBA{R: 100, G: 100, B: 100, A: 255},
			wantR: 100,
			wantG: 100,
			wantB: 100,
			wantA: 255,
		},
	} {
		t.Run(name, func(t *testing.T) {
			gotR, gotG, gotB, gotA := color.ToNRGBA(tt.color)
			assert.Equal(t, tt.wantR, gotR)
			assert.Equal(t, tt.wantG, gotG)
			assert.Equal(t, tt.wantB, gotB)
			assert.Equal(t, tt.wantA, gotA)
		})
	}
}
