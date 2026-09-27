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
type overrideMap struct{ m sync.Map } // weak.Pointer[byte] → *overrideScope

func overrideKey(o fyne.CanvasObject) (weak.Pointer[byte], bool) {
	v := reflect.ValueOf(o)
	if o == nil || v.Kind() != reflect.Pointer || v.IsNil() {
		return weak.Pointer[byte]{}, false
	}
	return weak.Make((*byte)(v.UnsafePointer())), true
}

func (m *overrideMap) Load(o fyne.CanvasObject) (*overrideScope, bool) {
	k, ok := overrideKey(o)
	if !ok {
		return nil, false
	}
	v, ok := m.m.Load(k)
	if !ok {
		return nil, false
	}
	return v.(*overrideScope), true
}

func (m *overrideMap) Store(o fyne.CanvasObject, s *overrideScope) {
	if k, ok := overrideKey(o); ok {
		m.m.Store(k, s)
	}
}

func (m *overrideMap) Delete(o fyne.CanvasObject) {
	if k, ok := overrideKey(o); ok {
		m.m.Delete(k)
	}
}
