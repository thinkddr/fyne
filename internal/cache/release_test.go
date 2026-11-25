package cache

import (
	"runtime"
	"testing"

	"fyne.io/fyne/v2"

	"github.com/stretchr/testify/assert"
)

func TestReleaseTree_FreesNowWhatWouldExpireLater(t *testing.T) {
	testClearAll()
	c := &dummyCanvas{}
	destroyed := 0
	rect := &dummyObject{}
	child := &dummyWidget{onDestroy: func() { destroyed++ }}
	parent := &dummyWidget{onDestroy: func() { destroyed++ }}
	kept := &dummyWidget{onDestroy: func() { destroyed++ }}
	Renderer(parent).(*dummyWidgetRenderer).objects = []fyne.CanvasObject{child, rect}
	Renderer(child)
	Renderer(kept)
	tree := &fyne.Container{Objects: []fyne.CanvasObject{parent, kept}}
	for _, o := range []fyne.CanvasObject{tree, parent, child, rect, kept} {
		SetCanvasForObject(o, c, nil)
	}
	SetTexture(rect, 1, c)
	SetTexture(kept, 2, c)

	k := Kept{}
	k.Keep(&fyne.Container{Objects: []fyne.CanvasObject{kept}})
	ReleaseTree(tree, k)

	assert.Equal(t, 2, destroyed)
	assert.False(t, IsRendered(parent))
	assert.False(t, IsRendered(child), "the walk goes down through cached renderers")
	assert.True(t, IsRendered(kept))
	assert.Nil(t, GetCanvasForObject(rect))
	assert.Equal(t, c, GetCanvasForObject(kept))
	var expired []fyne.CanvasObject
	RangeExpiredTexturesFor(c, func(o fyne.CanvasObject) { expired = append(expired, o) })
	assert.Equal(t, []fyne.CanvasObject{rect}, expired, "the textures of the tree are freed at the next clean, not a minute later")
}

func TestReleaseTree_CreatesNoRenderer(t *testing.T) {
	testClearAll()
	never := &dummyWidget{}
	ReleaseTree(&fyne.Container{Objects: []fyne.CanvasObject{never}}, nil)
	assert.False(t, IsRendered(never))
	assert.Equal(t, 0, renderers.Len())
}

// dummyObject is a canvas object that is neither a widget nor a container.
type dummyObject struct {
	fyne.CanvasObject
}

func TestReleaseTree_ForgetsTheScopeOfEveryObject(t *testing.T) {
	testClearAll()
	defer ResetThemeCaches()
	prim := &dummyObject{}
	wid := &dummyWidget{}
	Renderer(wid)
	inner := &fyne.Container{Objects: []fyne.CanvasObject{prim}}
	tree := &fyne.Container{Objects: []fyne.CanvasObject{inner, wid}}
	NewOverrideScope(nil).Apply(tree)
	for _, o := range []fyne.CanvasObject{prim, wid} {
		_, ok := overrides.Load(o)
		assert.True(t, ok)
	}

	ReleaseTree(tree, nil)
	for _, o := range []fyne.CanvasObject{tree, inner, prim, wid} {
		_, ok := overrides.Load(o)
		assert.False(t, ok, "a released object keeps the scope entry of the override it was under")
	}
}

func TestClean_DropsTheScopesOfCollectedObjects(t *testing.T) {
	testClearAll()
	defer ResetThemeCaches()
	overrides.m.Clear()
	scope := NewOverrideScope(nil)
	for range 100 {
		scope.Apply(&fyne.Container{Objects: []fyne.CanvasObject{&dummyObject{}}})
	}
	kept := &dummyObject{}
	scope.Apply(kept)
	runtime.GC()
	runtime.GC()

	tm := &timeMock{}
	tm.setTime(10, 0)
	lastClean = tm.createTime(9, 0)
	Clean(true)

	n := 0
	overrides.m.Range(func(any, any) bool { n++; return true })
	assert.Equal(t, 1, n, "the entries of collected objects are dropped on clean")
	_, ok := overrides.Load(kept)
	assert.True(t, ok)
	runtime.KeepAlive(kept)
}
