//go:build go1.24

package cache

import (
	"reflect"
	"sync"
	"weak"

	"fyne.io/fyne/v2"
)

// overrideMap maps a canvas object to its theme override scope without keeping the
// object alive. A plain map kept every object ever placed under a ThemeOverride for good:
// canvas primitives are never deleted (only widgets are, when their renderer expires),
// and a primitive's closures (a Raster generator, say) held the whole screen that built it.
type overrideMap struct{ m sync.Map } // weak.Pointer[byte] → *OverrideScope

func overrideKey(o fyne.CanvasObject) (weak.Pointer[byte], bool) {
	v := reflect.ValueOf(o)
	if o == nil || v.Kind() != reflect.Pointer || v.IsNil() {
		return weak.Pointer[byte]{}, false
	}
	return weak.Make((*byte)(v.UnsafePointer())), true
}

func (m *overrideMap) Load(o fyne.CanvasObject) (*OverrideScope, bool) {
	k, ok := overrideKey(o)
	if !ok {
		return nil, false
	}
	v, ok := m.m.Load(k)
	if !ok {
		return nil, false
	}
	return v.(*OverrideScope), true
}

func (m *overrideMap) Store(o fyne.CanvasObject, s *OverrideScope) {
	if k, ok := overrideKey(o); ok {
		m.m.Store(k, s)
	}
}

// StoreIfChanged stores s for o unless o already has it. Applying a scope again walks
// the whole tree, and most of it already belongs to the scope: storing it anyway costs
// an allocation per object for nothing.
func (m *overrideMap) StoreIfChanged(o fyne.CanvasObject, s *OverrideScope) {
	k, ok := overrideKey(o)
	if !ok {
		return
	}
	if v, ok := m.m.Load(k); ok && v.(*OverrideScope) == s {
		return
	}
	m.m.Store(k, s)
}

func (m *overrideMap) Delete(o fyne.CanvasObject) {
	if k, ok := overrideKey(o); ok {
		m.m.Delete(k)
	}
}

// dropDead deletes the entries of objects that have been collected. Their weak keys no
// longer match anything, but the entries stayed: canvas primitives are never deleted
// (only widgets are, when their renderer is destroyed), so a ThemeOverride whose content
// is rebuilt grew this map by one entry per primitive, for good.
func (m *overrideMap) dropDead() {
	m.m.Range(func(k, _ any) bool {
		if k.(weak.Pointer[byte]).Value() == nil {
			m.m.Delete(k)
		}
		return true
	})
}
