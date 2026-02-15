//go:build !noos && !tinygo

package app

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

func rootCacheDir(a fyne.App) string {
	desktopCache, _ := os.UserCacheDir()
	return filepath.Join(desktopCache, "fyne", a.UniqueID())
}
