package container

import (
	"reflect"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/cache"
	intTheme "fyne.io/fyne/v2/internal/theme"
	"fyne.io/fyne/v2/widget"
)

// ThemeOverride is a container where the child widgets are themed by the specified theme.
// Containers will be traversed and all child widgets will reflect the theme in this container.
// This should be used sparingly to avoid a jarring user experience.
//
// Since: 2.5
type ThemeOverride struct {
	widget.BaseWidget

	Content fyne.CanvasObject
	Theme   fyne.Theme

	holder *fyne.Container

	features *themeFeatures

	// scope is what this container applies to its content, created for scopeTheme.
	// It is kept while Theme does not change, so a Refresh walks the content again
	// for new objects without giving them a new cache ID: a new one would mean
	// rendering every themed icon again and resetting the theme caches of the app.
	scope      *cache.OverrideScope
	scopeTheme fyne.Theme
}

// NewThemeOverride provides a container where the child widgets are themed by the specified theme.
// Containers will be traversed and all child widgets will reflect the theme in this container.
// This should be used sparingly to avoid a jarring user experience.
//
// If the content `obj` of this theme override is a container and items are later added to the container or any
// sub-containers ensure that you call `Refresh()` on this `ThemeOverride` to ensure the new items match the theme.
//
// Since: 2.5
func NewThemeOverride(obj fyne.CanvasObject, th fyne.Theme) *ThemeOverride {
	t := &ThemeOverride{Content: obj, Theme: th, holder: NewStack(obj), features: &themeFeatures{}}
	t.ExtendBaseWidget(t)

	t.themeScope().Apply(obj)
	obj.Refresh() // required as the widgets passed in could have been initially rendered with default theme
	return t
}

func (t *ThemeOverride) CreateRenderer() fyne.WidgetRenderer {
	t.themeScope().Apply(t.Content)

	return &overrideRenderer{parent: t, objs: []fyne.CanvasObject{t.holder}}
}

func (t *ThemeOverride) Refresh() {
	if t.holder.Objects[0] != t.Content {
		t.holder.Objects[0] = t.Content
		t.holder.Refresh()
	}

	t.themeScope().Apply(t.Content)
	t.Content.Refresh()
	t.BaseWidget.Refresh()
}

// SetDeviceIsMobile allows a ThemeOverride container to shape the contained widgets as a mobile device.
// This will impact containers such as AppTabs and DocTabs, and more in the future, to display a layout
// that would automatically be used for a mobile device runtime.
//
// Since: 2.6
func (t *ThemeOverride) SetDeviceIsMobile(on bool) {
	t.features.mobile = on
	t.Refresh()
}

// themeScope returns the scope for the current Theme, creating a new one (and so a new
// cache ID) only when Theme is not the value the current scope was created for.
// A theme that is changed in place keeps its scope: assign a new value to Theme
// for its icons to be rendered again.
func (t *ThemeOverride) themeScope() *cache.OverrideScope {
	if t.scope == nil || !sameTheme(t.scopeTheme, t.Theme) {
		t.scope = cache.NewOverrideScope(addFeatures(t.Theme, t))
		t.scopeTheme = t.Theme
	}
	return t.scope
}

// sameTheme reports whether a and b are the same theme value. Values of a type that
// cannot be compared (a struct holding a map, say) never are, rather than panicking.
func sameTheme(a, b fyne.Theme) bool {
	ta := reflect.TypeOf(a)
	if ta != reflect.TypeOf(b) {
		return false
	}
	if ta != nil && !ta.Comparable() {
		return false
	}
	return a == b
}

// themeFeatures is what a ThemeOverride tells the themed objects about itself. The
// theme holds this and not the container: the theme is kept against every object it
// covers, and a pointer back to the container (and so to those same objects) kept them
// all alive for good.
type themeFeatures struct {
	mobile bool
}

type featureTheme struct {
	fyne.Theme

	over *themeFeatures
}

func addFeatures(th fyne.Theme, o *ThemeOverride) fyne.Theme {
	if o.features == nil {
		o.features = &themeFeatures{}
	}
	return &featureTheme{Theme: th, over: o.features}
}

func (f *featureTheme) Feature(n intTheme.FeatureName) any {
	if n == intTheme.FeatureNameDeviceIsMobile {
		return f.over.mobile
	}

	return nil
}

type overrideRenderer struct {
	parent *ThemeOverride

	objs []fyne.CanvasObject
}

func (r *overrideRenderer) Destroy() {
}

func (r *overrideRenderer) Layout(s fyne.Size) {
	intTheme.PushRenderingTheme(r.parent.Theme)
	defer intTheme.PopRenderingTheme()

	r.parent.holder.Resize(s)
}

func (r *overrideRenderer) MinSize() fyne.Size {
	intTheme.PushRenderingTheme(r.parent.Theme)
	defer intTheme.PopRenderingTheme()

	return r.parent.Content.MinSize()
}

func (r *overrideRenderer) Objects() []fyne.CanvasObject {
	return r.objs
}

func (r *overrideRenderer) Refresh() {
}
