package cache

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2"
)

// The theme caches belong to the whole app. A scope is applied to every object it
// covers, again on each refresh of a ThemeOverride and to each new row of a list
// under one, so applying it must not reset them: only creating a scope does, once.
func TestOverrideScope_ApplyKeepsThemeCaches(t *testing.T) {
	defer func() { testClearAll(); ResetThemeCaches() }()
	scope := NewOverrideScope(nil)
	tree, widgets := themedTree(50)
	seedThemeCaches()

	scope.Apply(tree)
	for _, w := range widgets {
		assert.Equal(t, scope.cacheID, WidgetScopeID(w))
	}
	assertThemeCachesSeeded(t, "applying a scope to 50 widgets")

	scope.Apply(tree)
	assertThemeCachesSeeded(t, "applying the scope again")

	row, _ := themedTree(1)
	assert.True(t, OverrideThemeMatchingScope(row, widgets[0]))
	assertThemeCachesSeeded(t, "a new row taking the scope of its list")

	other := NewOverrideScope(nil)
	assert.Nil(t, GetSvg("probe.svg", nil, 1, 1), "a new scope resets the caches")
	other.Apply(tree)
	for _, w := range widgets {
		assert.Equal(t, other.cacheID, WidgetScopeID(w), "a new scope replaces the old one")
	}
}

func TestOverrideTheme_NewScopeEachCall(t *testing.T) {
	defer func() { testClearAll(); ResetThemeCaches() }()
	tree, widgets := themedTree(2)

	OverrideTheme(tree, nil)
	first := WidgetScopeID(widgets[0])
	assert.NotEmpty(t, first)
	assert.Equal(t, first, WidgetScopeID(widgets[1]))

	seedThemeCaches()
	OverrideTheme(tree, nil)
	assert.NotEqual(t, first, WidgetScopeID(widgets[0]))
	assert.Nil(t, GetSvg("probe.svg", nil, 1, 1), "a new scope resets the caches")
}

// Applying a scope again stores nothing for the objects that already have it: a
// store allocates, once per object of the tree, on every refresh.
func TestOverrideScope_ApplyAgainDoesNotAllocate(t *testing.T) {
	defer func() { testClearAll(); ResetThemeCaches() }()
	scope := NewOverrideScope(nil)
	tree, _ := themedTree(20)
	scope.Apply(tree)

	assert.Zero(t, testing.AllocsPerRun(5, func() { scope.Apply(tree) }))
}

func BenchmarkOverrideScope_ApplyAgain(b *testing.B) {
	defer func() { testClearAll(); ResetThemeCaches() }()
	scope := NewOverrideScope(nil)
	tree, _ := themedTree(200)
	scope.Apply(tree)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		scope.Apply(tree)
	}
}

func seedThemeCaches() {
	SetSvg("probe.svg", nil, image.NewNRGBA(image.Rect(0, 0, 1, 1)), 1, 1)
	SetFontMetrics("probe", 12, fyne.TextStyle{}, nil, fyne.NewSize(10, 12), 9)
}

func assertThemeCachesSeeded(t *testing.T, after string) {
	t.Helper()
	assert.NotNil(t, GetSvg("probe.svg", nil, 1, 1), "SVG cache reset by %s", after)
	size, _ := GetFontMetrics("probe", 12, fyne.TextStyle{}, nil)
	assert.Equal(t, fyne.NewSize(10, 12), size, "text metrics reset by %s", after)
}

// themedTree is a container of n widgets, each holding a widget and a primitive.
func themedTree(n int) (*fyne.Container, []fyne.Widget) {
	tree := &fyne.Container{}
	var widgets []fyne.Widget
	for range n {
		inner := &parentWidget{children: []fyne.CanvasObject{&primitive{}}}
		outer := &parentWidget{children: []fyne.CanvasObject{inner, &primitive{}}}
		tree.Objects = append(tree.Objects, outer)
		widgets = append(widgets, outer, inner)
	}
	return tree, widgets
}

type parentWidget struct {
	fyne.Widget
	children []fyne.CanvasObject
}

func (w *parentWidget) CreateRenderer() fyne.WidgetRenderer {
	return &dummyWidgetRenderer{widget: &dummyWidget{}, objects: w.children}
}

type primitive struct {
	fyne.CanvasObject
	_ int // not zero-sized, so that each one has an address of its own
}
