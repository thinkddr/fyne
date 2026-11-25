package mobile

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/internal/driver/mobile/event/key"
)

// keyModifier maps the modifiers of a mobile key event to Fyne's.
func keyModifier(mod key.Modifiers) fyne.KeyModifier {
	var m fyne.KeyModifier
	if mod&key.ModShift != 0 {
		m |= fyne.KeyModifierShift
	}
	if mod&key.ModControl != 0 {
		m |= fyne.KeyModifierControl
	}
	if mod&key.ModAlt != 0 {
		m |= fyne.KeyModifierAlt
	}
	if mod&key.ModMeta != 0 {
		m |= fyne.KeyModifierSuper
	}
	return m
}

// keyShortcut is the shortcut that a key pressed on a hardware keyboard stands for,
// or nil if it is typed as usual. It follows the desktop driver: Control (Command on
// iOS) with Z, Y, V, C, X or A is the standard shortcut, and any other key held with
// Control, Alt or Meta is a desktop.CustomShortcut.
//
// The modifiers used to be ignored here, so Ctrl+Up reached the focused widget as a
// plain Up: a chat that edits your last message on Up did it on Ctrl+Up too. A key
// that still produces a character with Alt alone is typing, not a shortcut: AltGr,
// which Android reports as Alt, writes "@" or "€" on many European layouts.
func keyShortcut(keyName fyne.KeyName, r rune, mod key.Modifiers) fyne.Shortcut {
	if keyName == "" || mod&(key.ModControl|key.ModAlt|key.ModMeta) == 0 {
		return nil
	}
	if mod&(key.ModControl|key.ModMeta) == 0 && runeToPrintable(r) != 0 {
		return nil
	}
	modifier := keyModifier(mod)
	if modifier == fyne.KeyModifierShortcutDefault {
		switch keyName {
		case fyne.KeyZ:
			return &fyne.ShortcutUndo{}
		case fyne.KeyY:
			return &fyne.ShortcutRedo{}
		case fyne.KeyV:
			return &fyne.ShortcutPaste{Clipboard: NewClipboard()}
		case fyne.KeyC:
			return &fyne.ShortcutCopy{Clipboard: NewClipboard()}
		case fyne.KeyX:
			return &fyne.ShortcutCut{Clipboard: NewClipboard()}
		case fyne.KeyA:
			return &fyne.ShortcutSelectAll{}
		}
	}
	return &desktop.CustomShortcut{KeyName: keyName, Modifier: modifier}
}

// typeShortcut delivers a shortcut as the desktop driver does: to the focused
// object if it handles shortcuts - a disabled one only gets Copy - and to the canvas
// otherwise.
func (d *driver) typeShortcut(c *canvas, s fyne.Shortcut) {
	focused, ok := c.Focused().(fyne.Shortcutable)
	if !ok {
		c.TypedShortcut(s)
		return
	}
	type selectableText interface {
		fyne.Disableable
		SelectedText() string
	}
	if w, ok := focused.(selectableText); ok && w.Disabled() && s.ShortcutName() != "Copy" {
		return
	}
	focused.TypedShortcut(s)
}
