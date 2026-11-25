package app

import (
	"testing"

	"fyne.io/fyne/v2/internal/driver/mobile/event/key"
)

func TestKeyModifiers(t *testing.T) {
	for meta, want := range map[int32]key.Modifiers{
		0:                     0,
		0x01 | 0x40:           key.ModShift,   // AMETA_SHIFT_ON | AMETA_SHIFT_LEFT_ON
		0x1000 | 0x2000:       key.ModControl, // AMETA_CTRL_ON | AMETA_CTRL_LEFT_ON
		0x02 | 0x20:           key.ModAlt,     // AMETA_ALT_ON | AMETA_ALT_RIGHT_ON (AltGr)
		0x10000 | 0x20000:     key.ModMeta,    // AMETA_META_ON | AMETA_META_LEFT_ON
		0x1000 | 0x01:         key.ModControl | key.ModShift,
		0x100000 | 0x200000:   0, // AMETA_CAPS_LOCK_ON | AMETA_NUM_LOCK_ON are not modifiers
		0x1000 | 0x02 | 0x200: key.ModControl | key.ModAlt,
	} {
		if got := keyModifiers(meta); got != want {
			t.Errorf("meta %#x gives %v, want %v", meta, got, want)
		}
	}
}

func TestKeyIsShortcut(t *testing.T) {
	for _, c := range []struct {
		name string
		ev   key.Event
		want bool
	}{
		{"up", key.Event{Rune: -1, Code: key.CodeUpArrow}, false},
		{"ctrl+up", key.Event{Rune: -1, Code: key.CodeUpArrow, Modifiers: key.ModControl}, true},
		{"ctrl+v", key.Event{Rune: -1, Code: key.CodeV, Modifiers: key.ModControl}, true},
		{"meta+c", key.Event{Rune: -1, Code: key.CodeC, Modifiers: key.ModMeta}, true},
		{"alt+up", key.Event{Rune: -1, Code: key.CodeUpArrow, Modifiers: key.ModAlt}, true},
		{"altgr+2 types @", key.Event{Rune: '@', Code: key.Code2, Modifiers: key.ModAlt}, false},
		{"shift+a types A", key.Event{Rune: 'A', Code: key.CodeA, Modifiers: key.ModShift}, false},
		{"ctrl alone", key.Event{Rune: -1, Code: key.CodeLeftControl, Modifiers: key.ModControl}, false},
	} {
		if got := keyIsShortcut(c.ev); got != c.want {
			t.Errorf("%s: keyIsShortcut = %v, want %v", c.name, got, c.want)
		}
	}
}
