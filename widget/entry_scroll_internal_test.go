package widget

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fyne.io/fyne/v2"
	intwidget "fyne.io/fyne/v2/internal/widget"
	"fyne.io/fyne/v2/test"
)

func TestEntry_UndoRestoresFittingViewport(t *testing.T) {
	// Match the setup and capture order of TestEntry_UndoRedoImage.
	test.NewApp()
	t.Cleanup(func() { test.NewApp() })
	e := &Entry{MultiLine: true, Wrapping: fyne.TextWrapWord, Scroll: intwidget.ScrollNone}
	w := test.NewTempWindow(t, e)
	w.Resize(fyne.NewSize(150, 200))
	e.Resize(fyne.NewSize(120, 100))
	e.Move(fyne.NewPos(10, 10))
	test.AssertRendersToMarkup(t, "entry/initial_multiline.xml", w.Canvas())
	w.Resize(fyne.NewSize(128, 128))
	c := w.Canvas()
	c.Focus(e)

	logState := func(phase string) {
		provider := e.textProvider()
		content := e.scroll.Content
		t.Logf("%s: rows=%d cursor=(%d,%d) textOffset=%d cursorPosition=%v scrollSize=%v scrollOffset=%v contentPosition=%v contentSize=%v contentMinSize=%v providerSize=%v providerMinSize=%v",
			phase, provider.rows(), e.CursorRow, e.CursorColumn,
			textPosFromRowCol(e.CursorRow, e.CursorColumn, provider), e.CursorPosition(),
			e.scroll.Size(), e.scroll.Offset, content.Position(), content.Size(), content.MinSize(), provider.Size(), provider.MinSize())
	}
	encode := func(img image.Image) []byte {
		var buf bytes.Buffer
		require.NoError(t, png.Encode(&buf, img))
		return buf.Bytes()
	}

	const initialText = "The undo/\nredo function allows you to efficiently fix"
	for _, r := range initialText {
		e.TypedRune(r)
	}
	initial := encode(c.Capture())
	logState("initial")
	initialRow, initialColumn := e.CursorRow, e.CursorColumn
	assert.LessOrEqual(t, e.scroll.Content.MinSize().Height, e.scroll.Size().Height, "initial text fits vertically")
	assert.Zero(t, e.scroll.Offset.Y, "initial text needs no vertical scroll")

	for _, r := range " mistkaes" {
		e.TypedRune(r)
	}
	c.Capture()
	logState("overflow")
	assert.Greater(t, e.scroll.Content.MinSize().Height, e.scroll.Size().Height, "added text overflows vertically")
	assert.Greater(t, e.scroll.Offset.Y, float32(0), "the caret scrolls into view")

	e.TypedShortcut(&fyne.ShortcutUndo{})
	c.Capture()
	logState("undo")
	require.Equal(t, initialText, e.Text)
	assert.Equal(t, initialRow, e.CursorRow, "undo restores the cursor row")
	assert.Equal(t, initialColumn, e.CursorColumn, "undo restores the cursor column")
	assert.LessOrEqual(t, e.scroll.Content.MinSize().Height, e.scroll.Size().Height, "restored text fits vertically")
	if e.scroll.Content.MinSize().Height <= e.scroll.Size().Height {
		assert.Zero(t, e.scroll.Offset.Y, "undo must clear scrolling when the restored text fits")
	}

	// Probe whether refreshing the scroll layout repairs the recorded stale state.
	e.scroll.Refresh()
	corrected := encode(c.Capture())
	logState("undo-after-scroll-refresh")
	assert.Zero(t, e.scroll.Offset.Y, "refresh clamps the offset of fitting content")
	assert.Equal(t, e.scroll.Size(), e.scroll.Content.Size(), "fitting content uses the viewport size")
	assert.True(t, bytes.Equal(initial, corrected), "restored text and cursor paint the original viewport")
}
