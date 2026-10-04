# Fyne — Sytue fork

A downstream fork of [Fyne](https://github.com/fyne-io/fyne), the Go toolkit for
native desktop and mobile applications. The fork adds incoming URL handling,
Android multi-file selection, overlay/focus behavior and targeted rendering,
cache and lifecycle fixes.

**Latest tagged release:** `v2.8.1-sytue.16`, based on upstream `v2.8.1`.
The [initial fork delta](https://github.com/thinkddr/fyne/compare/v2.8.1...v2.8.1-sytue.15)
contains 17 commits through `v2.8.1-sytue.15`. The
[maintenance delta](https://github.com/thinkddr/fyne/compare/v2.8.1-sytue.15...v2.8.1-sytue.16)
covers the changes in `.16`. [FORK-NOTICE.md](FORK-NOTICE.md) records their scope.
The toolkit code and Sytue additions use Fyne's
[BSD 3-Clause license](LICENSE). Bundled fonts retain their
[own license notices](theme/font/).

The `.16` maintenance release updates documentation, pinned regression CI,
lint-safe refactoring and watcher-test synchronization. Reviewed visual fixtures
reflect the fork's existing position rounding and multiline Entry row spacing.
It also clears stale Entry scrolling when undo restores text that fits the
viewport.

## Install

Keep the canonical module and import path, **`fyne.io/fyne/v2`**. Select this fork
with a pinned replacement in your application's module:

```sh
go get fyne.io/fyne/v2@v2.8.1
go mod edit -replace=fyne.io/fyne/v2=github.com/thinkddr/fyne/v2@v2.8.1-sytue.16
go mod tidy
```

Add your Fyne imports before `go mod tidy`, which removes unused dependencies.
The module requires Go 1.22 or newer; desktop builds also need a C compiler and
the platform development libraries described in the
[Fyne setup guide](https://developer.fyne.io/started/).

Remove the replacement to return to upstream:

```sh
go mod edit -dropreplace=fyne.io/fyne/v2
go mod tidy
```

## Differences from upstream 2.8.1

| Area | Change | Code and regression tests |
| --- | --- | --- |
| Incoming URLs | Optional `fyne.URLHandler`; buffered delivery on the graphical context. Android intents, iOS URL/Universal Link callbacks, macOS URL/document Apple Events and desktop startup URL arguments. | [API](app.go), [queue tests](internal/urlhandler/urlhandler_test.go), [macOS bridge](app/incoming_urls_darwin.m), [iOS bridge](internal/driver/mobile/app/darwin_ios.m) |
| File selection | `dialog.ShowFileOpenMultiple`: Android selects several files; other platforms return at most one. Android content URIs can expose a provider-reported size. | [dialog API](dialog/file.go), [mobile picker](internal/driver/mobile/file.go), [Android URI](internal/driver/mobile/uri_android.go) |
| Overlays and focus | Opt-in `TapThrough() bool` and `FocusScope() fyne.CanvasObject` methods; `ScrollThrough() bool` is available in the GLFW desktop path. | [contract](internal/driver/overlay_through.go), [desktop tests](internal/driver/glfw/window_through_test.go), [mobile tap tests](internal/driver/mobile/canvas_through_test.go), [focus tests](internal/app/focus_scope_test.go) |
| Clipping and text | Nested clips intersect. Multiline Entry layout, selection, caret and hit testing use the theme's line spacing; single-line entries keep their previous spacing. | [clip tests](internal/driver/util_test.go), [Entry tests](widget/entry_linespacing_internal_test.go) |
| Rendering | Software positions round to the nearest pixel; sizes still round up. GL/ES separate alpha blending avoids translucent framebuffer edges. | [software tests](internal/painter/software/painter_test.go), [GL](internal/painter/gl/gl_core.go), [ES](internal/painter/gl/gl_es.go) |
| Theme/font retention | Theme override object keys use weak pointers on Go ≥1.24. Theme-font faces share a cache key based on style, name, backing slice address and length. | [retention test](container/theme_retain_test.go), [font tests](internal/painter/font_test.go) |
| Scheduling and lifecycle | Fix cache expiry after an idle interval, dispatch background test work on the test goroutine, queue pre-loop mobile work, adopt iOS scenes and correct Android key actions. | [cache tests](internal/cache/base_test.go), [test driver](test/driver_test.go), [mobile queue test](internal/driver/mobile/driver_test.go), [key tests](internal/driver/mobile/app/keydirection_test.go) |

## Receive a URL

```go
package main

import (
    "net/url"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/widget"
)

func main() {
    a := app.NewWithID("org.example.url-demo")
    w := a.NewWindow("Incoming URL")
    message := widget.NewLabel("Open an example:// URL")
    w.SetContent(message)
    if receiver, ok := a.(fyne.URLHandler); ok {
        receiver.SetOnOpenURL(func(u *url.URL) {
            if u.Scheme == "example" {
                message.SetText(u.String())
            }
        })
    }
    w.ShowAndRun()
}
```

On Linux or Windows, try `go run . 'example://hello'`. OS launches need your
application's scheme/document associations. Universal Links also need the iOS
entitlements and website association. The fork delivers URLs; the application
validates allowed routes and any authentication state. Linux/Windows handling is
startup-only; it does not forward URLs to an already running process.

For multi-file selection, use `dialog.ShowFileOpenMultiple(callback, window)`.
Cancellation returns a nil slice. Close every returned reader when finished.
Android provider sizes use an optional `interface{ Size() int64 }` assertion on
the URI; `-1` means unknown and this method is not part of `fyne.URI`.

## Scope and evidence

The linked tests cover their stated regressions. Headless tests do not prove
Android/iOS/macOS native integration or compositor behavior. Those paths require
device or platform validation.

[FORK-VALIDATION.md](FORK-VALIDATION.md) explains the rendering reference,
reviewed fixture updates and release checks.

The strict software-image reference uses Linux x86-64. The
[unchanged upstream 2.8.1 reference run](https://github.com/thinkddr/fyne/actions/runs/37198815282)
passes in that environment; Linux ARM64 produces image differences with the same
Go 1.27.1 toolchain. ARM64 browser or native pixel equivalence requires its own
verification.

Go 1.22/1.23 retain the strong theme-key fallback. Weak keys do not establish a
bound on all cache metadata. The theme-font cache is **not** content-addressed:
identical bytes in separate allocations do not share its identity. Explicit
custom font sources still require comparable `fyne.Resource` values.

Rendering fixes do not establish CSS support or browser pixel parity. The
[Astro Fyne converter](https://github.com/thinkddr/astro-fyne) maintains its own
supported-source contracts and visual comparisons.

For the general toolkit tutorial, packaging and examples, read the preserved
[upstream README](UPSTREAM-README.md) and
[Fyne developer documentation](https://developer.fyne.io/).
Report fork-specific issues in [this repository](https://github.com/thinkddr/fyne/issues).
