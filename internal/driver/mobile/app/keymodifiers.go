package app

import (
	"strconv"

	"fyne.io/fyne/v2/internal/driver/mobile/event/key"
)

// Android meta state flags (AMETA_* in android/input.h). The *_ON flags are set
// whenever either the left or the right key is down.
const (
	androidMetaShiftOn = 0x01
	androidMetaAltOn   = 0x02
	androidMetaCtrlOn  = 0x1000
	androidMetaMetaOn  = 0x10000
)

// keyModifiers maps the meta state of an Android key event to key modifiers.
//
// They used to be left unset ("TODO(crawshaw): set Modifiers"), so a key pressed
// with Ctrl, Alt or Meta on a hardware keyboard arrived as the bare key: Ctrl+Up
// was Up, and Ctrl+C was C.
func keyModifiers(meta int32) key.Modifiers {
	var m key.Modifiers
	if meta&androidMetaShiftOn != 0 {
		m |= key.ModShift
	}
	if meta&androidMetaCtrlOn != 0 {
		m |= key.ModControl
	}
	if meta&androidMetaAltOn != 0 {
		m |= key.ModAlt
	}
	if meta&androidMetaMetaOn != 0 {
		m |= key.ModMeta
	}
	return m
}

// isModifierKey reports whether the key is a modifier key itself.
func isModifierKey(c key.Code) bool {
	switch c {
	case key.CodeLeftControl, key.CodeRightControl, key.CodeLeftShift, key.CodeRightShift,
		key.CodeLeftAlt, key.CodeRightAlt, key.CodeLeftGUI, key.CodeRightGUI:
		return true
	}
	return false
}

// keyIsShortcut reports whether the driver delivers this key as a shortcut rather
// than typing it: it is held with Control or Meta, or with Alt without producing a
// character. A character produced with Alt alone is typing - AltGr, which Android
// reports as Alt, writes "@" or "€" on many European layouts.
//
// Android must not also hand such a key to the hidden EditText behind the soft
// keyboard: it would apply Ctrl+V or Ctrl+A to it, and a paste would arrive twice,
// once as the shortcut and once through the text watcher.
func keyIsShortcut(k key.Event) bool {
	if isModifierKey(k.Code) {
		return false
	}
	if k.Modifiers&(key.ModControl|key.ModMeta) != 0 {
		return true
	}
	return k.Modifiers&key.ModAlt != 0 && !strconv.IsPrint(k.Rune)
}
