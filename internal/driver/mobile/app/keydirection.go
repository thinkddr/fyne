package app

import "fyne.io/fyne/v2/internal/driver/mobile/event/key"

// Android key event actions (AKEY_EVENT_ACTION_* in android/input.h).
const (
	androidKeyActionDown = 0
	androidKeyActionUp   = 1
)

// keyDirection maps an Android key event action to a key direction.
//
// The actions are AKEY_EVENT_ACTION_DOWN (0) and AKEY_EVENT_ACTION_UP (1). They
// used to be compared with AKEY_STATE_DOWN (1) and AKEY_STATE_UP (0), which are key
// states, not actions, and have the opposite values: every release was delivered as a
// press. While the soft keyboard is up the IME takes the press and types it through
// the text watcher, and the release came back to us as a second press, so every
// character from a hardware keyboard (or `adb shell input text`) was typed twice.
func keyDirection(action int32) key.Direction {
	switch action {
	case androidKeyActionDown:
		return key.DirPress
	case androidKeyActionUp:
		return key.DirRelease
	default:
		return key.DirNone
	}
}
