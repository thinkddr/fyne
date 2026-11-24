//go:build !go1.24

package cache

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/async"
)

// overrideMap without weak pointers (before Go 1.24) keeps its objects alive.
type overrideMap struct {
	async.Map[fyne.CanvasObject, *OverrideScope]
}

// StoreIfChanged stores s for o unless o already has it.
func (m *overrideMap) StoreIfChanged(o fyne.CanvasObject, s *OverrideScope) {
	if old, ok := m.Load(o); ok && old == s {
		return
	}
	m.Store(o, s)
}
