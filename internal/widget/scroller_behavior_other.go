//go:build ci || !darwin || ios || iossimulator

package widget

func isScrollerPageOnTap() bool {
	return false
}

// scrollBarAlwaysVisible is true on the desktops without a macOS-style preference
// (all but macOS, and macOS under the ci tag). On iOS and Android
// the bars only show while scrolling (mobileScrollBars).
func scrollBarAlwaysVisible() bool {
	return !mobileScrollBars
}

func subscribeScrollerStyle(_ func()) uint64 { return 0 }
func unsubscribeScrollerStyle(_ uint64)      {}
