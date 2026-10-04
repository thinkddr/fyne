//go:build !android && !ios && !mobile

package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2"
)

func TestEnsureDir(t *testing.T) {
	tmpDir := testPath("fynetest")

	ensureDirExists(tmpDir)
	if st, err := os.Stat(tmpDir); err != nil || !st.IsDir() {
		t.Error("Could not ensure directory exists")
	}

	os.Remove(tmpDir)
}

func TestWatchSettings(t *testing.T) {
	settings := &settings{}
	listener := make(chan fyne.Settings, 1)
	settings.AddChangeListener(listener)

	settings.fileChanged() // simulate the settings file changing

	select {
	case <-listener:
	case <-time.After(100 * time.Millisecond):
		t.Error("Settings listener was not called")
	}
}

func TestWatchFile(t *testing.T) {
	path := testPath("fyne-temp-watch.txt")
	f, _ := os.Create(path)
	f.Close()
	defer os.Remove(path)

	called := make(chan any, 1)
	watcher := watchFile(path, func() {
		select {
		case called <- true:
		default:
		}
	})
	if !assert.NotNil(t, watcher, "Could not start watcher") {
		return
	}
	defer watcher.Close()
	file, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	file.WriteString(" ")
	file.Close()

	waitForFileWatcher(t, called)
}

func TestFileWatcher_FileDeleted(t *testing.T) {
	path := testPath("fyne-temp-watch.txt")
	f, _ := os.Create(path)
	f.Close()
	defer os.Remove(path)

	called := make(chan any, 1)
	watcher := watchFile(path, func() {
		select {
		case called <- true:
		default:
		}
	})
	if watcher == nil {
		assert.Fail(t, "Could not start watcher")
		return
	}

	defer watcher.Close()
	os.Remove(path)
	f, _ = os.Create(path)

	waitForFileWatcher(t, called)
	f.Close()
}

func waitForFileWatcher(t *testing.T, called <-chan any) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		// The test driver dispatches background callbacks on the test goroutine.
		fyne.DoAndWait(func() {})
		select {
		case <-called:
			return
		case <-deadline.C:
			t.Fatal("File watcher callback was not called")
		case <-tick.C:
		}
	}
}

func testPath(child string) string {
	// TMPDIR would be more normal but fsnotify cannot watch that on macOS...
	return filepath.Join("testdata", child)
}
