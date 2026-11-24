//go:build ci || no_glfw || android || ios || mobile

package widget

import "fyne.io/fyne/v2"

// DragEnd will stop scrolling on mobile has stopped.
// Where the bars only show while scrolling, they hide shortly after.
func (s *Scroll) DragEnd() {
	if s.scrolling {
		s.endScrollingAfterDelay()
	}
}

// Dragged will scroll on any drag - bar or otherwise - for mobile.
// Where the bars only show while scrolling, they show until the drag ends.
func (s *Scroll) Dragged(e *fyne.DragEvent) {
	shown := s.startScrolling()
	if s.updateOffset(e.Dragged.DX, e.Dragged.DY) || shown {
		s.refreshWithoutOffsetUpdate()
	}
}
