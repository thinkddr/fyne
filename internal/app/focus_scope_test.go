package app_test

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/internal/app"
	"fyne.io/fyne/v2/widget"

	"github.com/stretchr/testify/assert"
)

// scopedStack keeps focus traversal inside `scope` while it is set.
type scopedStack struct {
	widget.BaseWidget
	content, scope fyne.CanvasObject
}

func (s *scopedStack) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.content)
}

func (s *scopedStack) FocusScope() fyne.CanvasObject { return s.scope }

func TestFocusManager_FocusScope(t *testing.T) {
	outside, inside1, inside2 := &focusable{}, &focusable{}, &focusable{}
	drawer := container.NewVBox(inside1, inside2)
	root := &scopedStack{content: container.NewVBox(outside, drawer)}
	root.ExtendBaseWidget(root)
	m := app.NewFocusManager(root)

	m.FocusNext()
	assert.Equal(t, outside, m.Focused(), "without a scope, traversal walks the whole content")

	root.scope = drawer
	m.FocusNext()
	assert.Equal(t, inside1, m.Focused(), "with a scope, Tab jumps into it")
	m.FocusNext()
	assert.Equal(t, inside2, m.Focused())
	m.FocusNext()
	assert.Equal(t, inside1, m.Focused(), "and wraps around inside it")
	m.FocusPrevious()
	assert.Equal(t, inside2, m.Focused(), "backwards too")

	root.scope = nil
	m.FocusNext()
	assert.Equal(t, outside, m.Focused(), "a scope that is released gives the content back")
}
