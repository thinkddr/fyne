//go:build !windows || !ci

package mobile

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/internal/driver/mobile/event/key"
	"fyne.io/fyne/v2/widget"

	"github.com/stretchr/testify/assert"
)

// keyRecorder is a focusable widget that records what the driver types into it.
type keyRecorder struct {
	widget.BaseWidget

	keys      []fyne.KeyName
	runes     []rune
	shortcuts []fyne.Shortcut
}

var (
	_ fyne.Focusable    = (*keyRecorder)(nil)
	_ fyne.Shortcutable = (*keyRecorder)(nil)
)

func newKeyRecorder() *keyRecorder {
	r := &keyRecorder{}
	r.ExtendBaseWidget(r)
	return r
}

func (r *keyRecorder) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(widget.NewLabel(""))
}
func (r *keyRecorder) FocusGained()                  {}
func (r *keyRecorder) FocusLost()                    {}
func (r *keyRecorder) TypedKey(e *fyne.KeyEvent)     { r.keys = append(r.keys, e.Name) }
func (r *keyRecorder) TypedRune(c rune)              { r.runes = append(r.runes, c) }
func (r *keyRecorder) TypedShortcut(s fyne.Shortcut) { r.shortcuts = append(r.shortcuts, s) }
func (r *keyRecorder) reset()                        { r.keys, r.runes, r.shortcuts = nil, nil, nil }

// A hardware keyboard on a phone or tablet: a key held with Control, Alt or Meta is a
// shortcut, as on the desktop, and never the bare key. Ctrl+Up used to reach the
// focused widget as a plain Up.
func TestTypeDownCanvas_ModifiersMakeShortcuts(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	rec := newKeyRecorder()
	c.SetContent(rec)

	d.DoFromGoroutine(func() {
		c.Focus(rec)

		// A plain Up is a key, as before.
		d.typeDownCanvas(c, -1, key.CodeUpArrow, 0)
		assert.Equal(t, []fyne.KeyName{fyne.KeyUp}, rec.keys)
		assert.Empty(t, rec.shortcuts)
		rec.reset()

		// Ctrl+Up is a shortcut, never a plain Up.
		d.typeDownCanvas(c, -1, key.CodeUpArrow, key.ModControl)
		assert.Empty(t, rec.keys, "Ctrl+Up must not arrive as Up")
		if assert.Len(t, rec.shortcuts, 1) {
			assert.Equal(t, &desktop.CustomShortcut{KeyName: fyne.KeyUp, Modifier: fyne.KeyModifierControl}, rec.shortcuts[0])
		}
		rec.reset()

		// Alt+Up and Ctrl+Shift+Up too, with all their modifiers.
		d.typeDownCanvas(c, -1, key.CodeUpArrow, key.ModAlt)
		d.typeDownCanvas(c, -1, key.CodeUpArrow, key.ModControl|key.ModShift)
		assert.Empty(t, rec.keys)
		if assert.Len(t, rec.shortcuts, 2) {
			assert.Equal(t, &desktop.CustomShortcut{KeyName: fyne.KeyUp, Modifier: fyne.KeyModifierAlt}, rec.shortcuts[0])
			assert.Equal(t, &desktop.CustomShortcut{KeyName: fyne.KeyUp, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift}, rec.shortcuts[1])
		}
		rec.reset()

		// The standard shortcuts, with the default modifier of the platform.
		def := key.ModControl
		if fyne.KeyModifierShortcutDefault == fyne.KeyModifierSuper {
			def = key.ModMeta
		}
		d.typeDownCanvas(c, -1, key.CodeC, def)
		d.typeDownCanvas(c, -1, key.CodeV, def)
		d.typeDownCanvas(c, -1, key.CodeA, def)
		if assert.Len(t, rec.shortcuts, 3) {
			assert.IsType(t, &fyne.ShortcutCopy{}, rec.shortcuts[0])
			assert.IsType(t, &fyne.ShortcutPaste{}, rec.shortcuts[1])
			assert.IsType(t, &fyne.ShortcutSelectAll{}, rec.shortcuts[2])
		}
		assert.Empty(t, rec.keys)
		assert.Empty(t, rec.runes)
		rec.reset()

		// Shift alone types, and AltGr (Alt with a character) types its character.
		d.typeDownCanvas(c, 'A', key.CodeA, key.ModShift)
		d.typeDownCanvas(c, '@', key.Code2, key.ModAlt)
		assert.Equal(t, []rune{'A', '@'}, rec.runes)
		assert.Empty(t, rec.shortcuts)
	}, true)
}

// With nothing focused the shortcut goes to the canvas, and the bare key does not.
func TestTypeDownCanvas_ShortcutWithoutFocusGoesToTheCanvas(t *testing.T) {
	c := newCanvas(fyne.CurrentDevice()).(*canvas)
	c.SetContent(widget.NewLabel("nothing focused"))
	var got fyne.Shortcut
	want := &desktop.CustomShortcut{KeyName: fyne.KeyUp, Modifier: fyne.KeyModifierControl}
	c.AddShortcut(want, func(s fyne.Shortcut) { got = s })
	typed := false
	c.SetOnTypedKey(func(*fyne.KeyEvent) { typed = true })

	d.DoFromGoroutine(func() {
		d.typeDownCanvas(c, -1, key.CodeUpArrow, key.ModControl)
	}, true)
	assert.Equal(t, want, got)
	assert.False(t, typed, "Ctrl+Up must not arrive as Up")
}
