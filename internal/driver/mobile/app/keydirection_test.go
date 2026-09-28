package app

import (
	"testing"

	"fyne.io/fyne/v2/internal/driver/mobile/event/key"
)

func TestKeyDirection(t *testing.T) {
	for action, want := range map[int32]key.Direction{
		androidKeyActionDown: key.DirPress,
		androidKeyActionUp:   key.DirRelease,
		2:                    key.DirNone, // AKEY_EVENT_ACTION_MULTIPLE
	} {
		if got := keyDirection(action); got != want {
			t.Errorf("action %d gives %v, want %v", action, got, want)
		}
	}
}
