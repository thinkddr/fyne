//go:build !go1.24

package cache

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/internal/async"
)

// overrideMap without weak pointers (before Go 1.24) keeps its objects alive.
type overrideMap struct {
	async.Map[fyne.CanvasObject, *overrideScope]
}
