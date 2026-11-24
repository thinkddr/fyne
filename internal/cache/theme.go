package cache

import (
	"strconv"
	"sync/atomic"

	"fyne.io/fyne/v2"
)

var (
	overrides     overrideMap
	overrideCount atomic.Uint32
)

// OverrideScope is a theme applied to a tree of objects. Its cache ID keys the theme
// caches of the objects it covers, so a scope can be applied again (to new or refreshed
// content) without invalidating anything rendered under it before.
type OverrideScope struct {
	th      fyne.Theme
	cacheID string
}

// NewOverrideScope creates a scope for th with a cache ID of its own.
//
// The theme caches are reset here, once per new scope, and not for each object it
// covers: applying a scope does not touch the caches, so walking a tree with many
// widgets, or walking it again on every refresh, keeps what the app has already
// rendered.
func NewOverrideScope(th fyne.Theme) *OverrideScope {
	id := overrideCount.Add(1)
	ResetThemeCaches()
	return &OverrideScope{th: th, cacheID: strconv.Itoa(int(id))}
}

// Apply makes o, and anything it contains, use the theme of this scope.
// Objects that already belong to this scope are left as they are.
func (s *OverrideScope) Apply(o fyne.CanvasObject) {
	overrideTheme(o, s)
}

// OverrideTheme allows an app to specify that a single object should use a different theme to the app.
// This should be used sparingly to avoid a jarring user experience.
// If the object is a container it will theme the children, if it is a canvas primitive it will do nothing.
//
// Each call creates a new scope; to theme the same tree again with the same theme,
// keep the scope from NewOverrideScope and Apply it instead.
//
// Since: 2.5
func OverrideTheme(o fyne.CanvasObject, th fyne.Theme) {
	NewOverrideScope(th).Apply(o)
}

func OverrideThemeMatchingScope(o, parent fyne.CanvasObject) bool {
	scope, ok := overrides.Load(parent)
	if !ok { // not overridden in parent
		return false
	}

	overrideTheme(o, scope)
	return true
}

func WidgetScopeID(o fyne.CanvasObject) string {
	scope, ok := overrides.Load(o)
	if !ok {
		return ""
	}

	return scope.cacheID
}

func WidgetTheme(o fyne.CanvasObject) fyne.Theme {
	scope, ok := overrides.Load(o)
	if !ok {
		return nil
	}

	return scope.th
}

func overrideContainer(c *fyne.Container, s *OverrideScope) {
	for _, o := range c.Objects {
		overrideTheme(o, s)
	}
}

func overrideTheme(o fyne.CanvasObject, s *OverrideScope) {
	if _, ok := o.(interface{ SetDeviceIsMobile(bool) }); ok { // ThemeOverride without the import loop
		return // do not apply this theme over a new scope
	}

	switch c := o.(type) {
	case fyne.Widget:
		overrideWidget(c, s)
	case *fyne.Container:
		overrideContainer(c, s)
	default:
		overrides.StoreIfChanged(c, s)
	}
}

func overrideWidget(w fyne.Widget, s *OverrideScope) {
	overrides.StoreIfChanged(w, s)

	r := Renderer(w)
	if r == nil {
		return
	}

	for _, o := range r.Objects() {
		overrideTheme(o, s)
	}
}
