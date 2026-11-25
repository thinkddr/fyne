package cache

import (
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
