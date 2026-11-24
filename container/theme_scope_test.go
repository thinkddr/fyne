package container

import (
	"image"
	"image/color"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/cache"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// A ThemeOverride walks its content again on every Refresh, and a list under one
// themes each row it creates. While the theme is the same, neither may give the
// content a new cache ID nor reset the theme caches of the whole app: both made every
// themed icon and text measure again, on every refresh of a themed list.
func TestThemeOverride_RefreshKeepsScopeAndThemeCaches(t *testing.T) {
	test.NewTempApp(t)
	rows := 3
	var icons []*widget.Icon
	list := widget.NewList(
		func() int { return rows },
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.MailComposeIcon())
			icons = append(icons, icon)
			return NewHBox(icon, widget.NewLabel(""))
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			o.(*fyne.Container).Objects[1].(*widget.Label).SetText(strconv.Itoa(id))
		})
	o := NewThemeOverride(list, test.Theme())
	w := test.NewTempWindow(t, o)
	w.Resize(fyne.NewSize(200, 300))
	w.Canvas().Capture()

	scope := cache.WidgetScopeID(list)
	require.NotEmpty(t, scope)
	created := len(icons)
	cache.SetSvg("probe.svg", nil, image.NewNRGBA(image.Rect(0, 0, 1, 1)), 1, 1)
	cache.SetFontMetrics("probe", 12, fyne.TextStyle{}, nil, fyne.NewSize(10, 12), 9)

	rows = 60
	o.Refresh()
	list.Refresh()
	list.ScrollToBottom()
	w.Canvas().Capture()
	o.Refresh()
	w.Canvas().Capture()

	assert.Equal(t, scope, cache.WidgetScopeID(list), "same theme, same scope")
	require.Greater(t, len(icons), created, "the list should create rows")
	for _, icon := range icons {
		assert.Equal(t, scope, cache.WidgetScopeID(icon))
	}
	assert.NotNil(t, cache.GetSvg("probe.svg", nil, 1, 1), "the SVG cache was reset")
	size, _ := cache.GetFontMetrics("probe", 12, fyne.TextStyle{}, nil)
	assert.Equal(t, fyne.NewSize(10, 12), size, "the text metrics cache was reset")
}

// Assigning another theme still gives the content a new scope, and its icons are
// rendered again in the new colours.
func TestThemeOverride_ThemeChangeRendersIconsAgain(t *testing.T) {
	test.NewTempApp(t)
	red := color.NRGBA{R: 0xff, A: 0xff}
	blue := color.NRGBA{B: 0xff, A: 0xff}
	button := widget.NewButtonWithIcon("", theme.HomeIcon(), nil)
	o := NewThemeOverride(button, tintTheme{test.Theme(), red})
	w := test.NewTempWindow(t, o)
	w.SetPadded(false)
	w.Resize(fyne.NewSize(40, 40))
	img := w.Canvas().Capture()
	assert.Positive(t, countPixels(img, red))
	assert.Zero(t, countPixels(img, blue))
	first := cache.WidgetScopeID(button)

	o.Theme = tintTheme{test.Theme(), red} // an equal value is the same theme
	o.Refresh()
	assert.Equal(t, first, cache.WidgetScopeID(button))

	cache.SetSvg("probe.svg", nil, image.NewNRGBA(image.Rect(0, 0, 1, 1)), 1, 1)
	o.Theme = tintTheme{test.Theme(), blue}
	o.Refresh()
	assert.NotEqual(t, first, cache.WidgetScopeID(button), "another theme, another scope")
	assert.Nil(t, cache.GetSvg("probe.svg", nil, 1, 1), "a new scope resets the caches")
	img = w.Canvas().Capture()
	assert.Zero(t, countPixels(img, red))
	assert.Positive(t, countPixels(img, blue))

	// A theme value that cannot be compared gets a new scope each time, as before.
	o.Theme = mapTheme{Theme: test.Theme(), m: map[string]int{}}
	o.Refresh()
	second := cache.WidgetScopeID(button)
	o.Refresh()
	assert.NotEqual(t, second, cache.WidgetScopeID(button))
}

type tintTheme struct {
	fyne.Theme
	fg color.Color
}

func (t tintTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameForeground {
		return t.fg
	}
	return t.Theme.Color(n, v)
}

type mapTheme struct {
	fyne.Theme
	m map[string]int
}

func countPixels(img image.Image, c color.NRGBA) (n int) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if color.NRGBAModel.Convert(img.At(x, y)) == c {
				n++
			}
		}
	}
	return n
}
