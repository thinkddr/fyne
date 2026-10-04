package widget

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/internal/cache"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

type lineSpacingTheme struct {
	fyne.Theme
	spacing float32
}

func (l lineSpacingTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNameLineSpacing {
		return l.spacing
	}
	return l.Theme.Size(n)
}

func spacedEntry(t *testing.T, multi bool, spacing float32) *Entry {
	e := NewEntry()
	if multi {
		e = NewMultiLineEntry()
	}
	w := test.NewTempWindow(t, e)
	cache.OverrideTheme(e, lineSpacingTheme{test.Theme(), spacing})
	if multi {
		e.SetText("one\ntwo\nthree")
	} else {
		e.SetText("one")
	}
	w.Resize(fyne.NewSize(200, 200))
	e.Refresh()
	return e
}

// A multi-line Entry steps its rows by the font height plus the theme's line spacing,
// as a CSS line-height does, and the cursor, the selection and a tap follow the rows.
func TestEntry_MultiLineRowsFollowLineSpacing(t *testing.T) {
	const gap = 10
	e := spacedEntry(t, true, gap)
	th := e.Theme()
	charH := e.text.charMinSize(false, e.TextStyle, th.Size(theme.SizeNameText)).Height

	texts := richTextRenderTexts(&e.text)
	assert.Len(t, texts, 3)
	for i := 1; i < 3; i++ {
		assert.InDelta(t, charH+gap, texts[i].Position().Y-texts[i-1].Position().Y, 0.01, "row %d", i)
	}

	flat := spacedEntry(t, true, 0)
	assert.InDelta(t, 2*gap, e.text.MinSize().Height-flat.text.MinSize().Height, 0.01, "text min size")
	assert.InDelta(t, 2*gap, e.MinSize().Height-flat.MinSize().Height, 0.01, "entry min size")

	var cursorY [3]float32
	for row := 0; row < 3; row++ {
		e.CursorRow, e.CursorColumn = row, 1
		cursorY[row] = e.CursorPosition().Y
		if row > 0 {
			assert.InDelta(t, texts[row].Position().Y-texts[0].Position().Y, cursorY[row]-cursorY[0], 0.01, "cursor row %d", row)
		}
		// A tap on the middle of the row, and one in the gap just below it, land on it.
		p := fyne.NewPos(e.CursorPosition().X+2, cursorY[row]+charH/2)
		r, _ := e.sel.getRowCol(p)
		assert.Equal(t, row, r, "tap in row %d", row)
		r, _ = e.sel.getRowCol(p.AddXY(0, charH/2+gap/2-1))
		assert.Equal(t, row, r, "tap in the gap under row %d", row)
	}

	// Select from the start of row 0 to the end of row 2: one box per row, at the row.
	e.CursorRow, e.CursorColumn = 0, 0
	e.sel.cursorRow, e.sel.cursorColumn = 0, 0
	e.KeyDown(&fyne.KeyEvent{Name: desktop.KeyShiftLeft})
	e.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	e.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	e.KeyUp(&fyne.KeyEvent{Name: desktop.KeyShiftLeft})
	boxes := test.TempWidgetRenderer(t, e.sel).(*selectableRenderer).selections
	assert.Len(t, boxes, 3)
	for i, b := range boxes {
		assert.InDelta(t, cursorY[i], b.Position().Y, 1.01, "selection row %d", i)
	}
}

// One line has no gap to add: a single-line Entry measures and paints the same whatever
// the line spacing.
func TestEntry_SingleLineIgnoresLineSpacing(t *testing.T) {
	a, b := spacedEntry(t, false, 0), spacedEntry(t, false, 10)
	assert.Equal(t, a.MinSize(), b.MinSize())
	assert.Equal(t, richTextRenderTexts(&a.text)[0].Position(), richTextRenderTexts(&b.text)[0].Position())
	assert.Equal(t, a.CursorPosition(), b.CursorPosition())
}
