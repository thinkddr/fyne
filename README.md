# Fyne — Sytue fork

A downstream fork of [Fyne](https://github.com/fyne-io/fyne), the Go toolkit for
native desktop and mobile applications. The fork adds incoming URL handling,
Android multi-file selection, overlay/focus behavior and targeted rendering,
cache and lifecycle fixes. It also provides opt-in shaped-text and rectangle
hooks for the software painter and optional mobile touch recognition.

**Version:** `v2.8.1-sytue.21`, based on upstream `v2.8.1`.
Available releases are listed in [Tags](https://github.com/thinkddr/fyne/tags).
The [initial fork delta](https://github.com/thinkddr/fyne/compare/v2.8.1...v2.8.1-sytue.15)
contains 17 commits through `v2.8.1-sytue.15`. The
[maintenance delta](https://github.com/thinkddr/fyne/compare/v2.8.1-sytue.15...v2.8.1-sytue.16)
covers the changes in `.16`. The
[software-image delta](https://github.com/thinkddr/fyne/compare/v2.8.1-sytue.16...v2.8.1-sytue.17)
describes `.17`. The
[cache maintenance delta](https://github.com/thinkddr/fyne/compare/v2.8.1-sytue.17...v2.8.1-sytue.18)
describes `.18`. The
[shaped-text delta](https://github.com/thinkddr/fyne/compare/v2.8.1-sytue.18...v2.8.1-sytue.19)
describes `.19`. [FORK-NOTICE.md](FORK-NOTICE.md) records their scope and the
`.20` rectangle-paint source and `.21` mobile touch addition.
The original toolkit uses Fyne's [BSD 3-Clause license](LICENSE). Fork changes
and additions use [Apache License 2.0](LICENSE-MODIFICATIONS). Bundled fonts retain their
[own license notices](theme/font/).

The `.16` maintenance release updates documentation, pinned regression CI,
lint-safe refactoring and watcher-test synchronization. Reviewed visual fixtures
reflect the fork's existing position rounding and multiline Entry row spacing.
It also clears stale Entry scrolling when undo restores text that fits the
viewport.

## Cache expiry in .18

The `.18` release synchronizes cache expiry timestamps when entry refresh and
cleanup overlap. It preserves the existing clock sample, duration calculation
and strict expiration boundary. A concurrent regression checks refreshes from
two writers alongside expiry reads, including the exact boundary.

Fresh main CI exposed this race after `.17` was published. The same-tree tag
checks passed, and the new regression reproduced the race against the original
implementation. See [the diagnosis and evidence](FORK-NOTICE.md#maintenance-release-v281-sytue18).

## Software image enlargement in .17

The `.17` release adds a bounded sampler for Smooth software Image objects.
It uses Skia-derived Q4 interpolation with floor quantization of premultiplied
channels. It preserves Raster, Pixels and Fastest routes and the original
CatmullRom fallback outside the new predicate. The image prepass is unchanged.

Eligibility is checked after that prepass: a zero-origin NRGBA8 or opaque RGBA8
buffer, uniform physical enlargement by exactly 2 or 4, object alpha 1 and no
corner radius. Original image types or bounds may be normalized by the prepass;
they are not an original-source exclusion guarantee. See
[FORK-NOTICE.md](FORK-NOTICE.md#software-image-release-v281-sytue17) for details.

[Published-module run 38900188638](https://github.com/thinkddr/astro-fyne/actions/runs/38900188638)
passed all six strict full-frame comparisons against the actual `.17` module for
the retained PNG8 corpus at DPR 1 and 2. It covered forward conversion, inverse
conversion and native reconstruction. The `.17` tag also passed all five fork
workflows and a clean public install. Its same-tree main run exposed the cache
race addressed in `.18`.

The final `.18` release head must pass all five fork workflows before
publication. A clean install and fresh strict comparisons against the published
`.18` module remain required; the `.17` evidence does not certify `.18` pixels.
See [the evidence details](FORK-NOTICE.md#release-checks-and-evidence).

## Install

Keep the canonical module and import path, **`fyne.io/fyne/v2`**. Select this fork
with a pinned replacement in your application's module:

```sh
go get fyne.io/fyne/v2@v2.8.1
go mod edit -replace=fyne.io/fyne/v2=github.com/thinkddr/fyne/v2@v2.8.1-sytue.21
go mod tidy
```

These commands require the published `.21` tag; see
[Tags](https://github.com/thinkddr/fyne/tags) for available versions.
Add your Fyne imports before `go mod tidy`, which removes unused dependencies.
The module requires Go 1.22 or newer; desktop builds also need a C compiler and
the platform development libraries described in the
[Fyne setup guide](https://developer.fyne.io/started/).

Remove the replacement to return to upstream:

```sh
go mod edit -dropreplace=fyne.io/fyne/v2
go mod tidy
```

## Opt-in shaped text in .19

Version `.19` adds [`canvas.Text.Layout`](canvas/text_layout.go). It keeps
text as a real canvas primitive while allowing a provider to supply an immutable
shaped line.
Its metrics define the line-box size and baseline. The provider owns the glyph
snapshot and any font engine; Fyne does not close native resources.
The hook adds no CGo or font-engine dependency to Fyne; an application chooses
and supplies its provider.

The software painter passes the current accumulated `*image.NRGBA`, restricted
to the intersected physical clip, with its real stride and backdrop. The context
contains the aligned, unrounded logical origin, actual canvas scale and resolved
text color. Providers paint glyphs only, respect the clip and return unsupported
requests before changing pixels. Fyne cannot undo partial provider writes.

Underline and strikethrough use the supplied advance and baseline with Fyne's
existing line stroke and offsets. They do not implement general CSS decoration
rules. Decoration rasters have coordinate checks and a 16 Mi-pixel allocation
limit, checked before the provider paints.

`LayoutMetrics` validates supplied metrics; invalid metrics return zero size
without default measurement. `TextLayoutError` exposes metric and paint errors.
A successful metric query does not clear a paint failure; successful software
painting or `SetTextLayoutError(nil)` clears the paint error. GL reports
`ErrTextLayoutUnsupportedGL` and draws no fallback glyphs. A nil layout preserves
the existing text path on every platform. This hook alone does not establish
browser, Fyne widget, editor, inverse-conversion or GPU parity.

The code at `eb30b4793d6eb857ae080acaa24d80a4c4ea6c19` passed all five fork
workflows. Those branch runs skipped the tag-only public-install check.
Release validation reruns all five workflows for the final source and tag, and
requires a clean installation of the actual published module. Results are
recorded in [Actions](https://github.com/thinkddr/fyne/actions). Consumer-specific
native text proof remains separate: the hook tests do not certify a font engine
or browser rendering profile. See the
[evidence and limits](FORK-NOTICE.md#software-shaped-text-in-v281-sytue19).

## Supplied rectangle painting in .20

`canvas.Rectangle.Painter` lets an application provide an immutable native paint
plan through `canvas.RectanglePainter`. Software painting receives the actual
accumulated `*image.NRGBA`, its original stride, the unrounded global origin,
logical size, canvas scale and intersected physical ancestor clip. The clip
allows outline outsets beyond the rectangle's own box. The provider owns its
resources and must validate unsupported requests before writing pixels.

A nil painter preserves the default rectangle path. A supplied painter replaces
that path completely; its error is exposed by `PaintError` without fallback.
Measurement and refresh retain the error. A successful paint or explicit
`SetPaintError(nil)` clears it. GL reports `ErrRectanglePainterUnsupportedGL`.
The hook adds no native library dependency to Fyne and does not itself implement
CSS outlines, browser composition, text shaping or GPU parity.

The implementation at `a82d070eff48adf2845f7887db7c82bb0d4488d2` passed all five
fork workflows, with thirteen successful jobs and the tag-only public-install
job skipped. Final-source and published-module checks are recorded separately;
see the [scope and evidence](FORK-NOTICE.md#supplied-rectangle-painting-in-v281-sytue20).

## Differences from upstream 2.8.1

| Area | Change | Code and regression tests |
| --- | --- | --- |
| Incoming URLs | Optional `fyne.URLHandler`; buffered delivery on the graphical context. Android intents, iOS URL/Universal Link callbacks, macOS URL/document Apple Events and desktop startup URL arguments. | [API](app.go), [queue tests](internal/urlhandler/urlhandler_test.go), [macOS bridge](app/incoming_urls_darwin.m), [iOS bridge](internal/driver/mobile/app/darwin_ios.m) |
| File selection | `dialog.ShowFileOpenMultiple`: Android selects several files; other platforms return at most one. Android content URIs can expose a provider-reported size. | [dialog API](dialog/file.go), [mobile picker](internal/driver/mobile/file.go), [Android URI](internal/driver/mobile/uri_android.go) |
| Overlays and focus | Opt-in `TapThrough() bool` and `FocusScope() fyne.CanvasObject` methods; `ScrollThrough() bool` is available in the GLFW desktop path. | [contract](internal/driver/overlay_through.go), [desktop tests](internal/driver/glfw/window_through_test.go), [mobile tap tests](internal/driver/mobile/canvas_through_test.go), [focus tests](internal/app/focus_scope_test.go) |
| Clipping and text | Nested clips intersect. Multiline Entry layout, selection, caret and hit testing use the theme's line spacing; single-line entries keep their previous spacing. | [clip tests](internal/driver/util_test.go), [Entry tests](widget/entry_linespacing_internal_test.go) |
| Rendering | Software positions round to the nearest pixel; sizes still round up. GL/ES separate alpha blending avoids translucent framebuffer edges. | [software tests](internal/painter/software/painter_test.go), [GL](internal/painter/gl/gl_core.go), [ES](internal/painter/gl/gl_es.go) |
| Image enlargement (.17) | Bounded post-prepass Smooth sampling for uniform physical factors 2 and 4; other cases retain the original fallback. | [sampler](internal/painter/software/skia_legacy_enlargement.go), [arithmetic tests](internal/painter/software/skia_legacy_enlargement_test.go), [Image/Raster routing](internal/painter/software/skia_legacy_image_test.go), [declined profiles](internal/painter/software/skia_legacy_scope_test.go) |
| Cache expiry (.18) | Atomic timestamps synchronize concurrent refresh and expiry checks, preserving the existing clock and strict boundary. | [expiry logic](internal/cache/base.go), [concurrent regression](internal/cache/expiry_concurrent_test.go), [expiry tests](internal/cache/base_test.go) |
| Shaped text (.19) | Opt-in immutable line metrics and software glyph painting into the real accumulated destination; a nil layout retains the default path and GL rejects supplied layouts explicitly. | [API and errors](canvas/text_layout.go), [metric tests](canvas/text_layout_test.go), [software controls](internal/painter/software/text_layout_test.go), [real canvas/clip tests](internal/painter/software/text_layout_integration_test.go), [GL rejection tests](internal/painter/gl/text_layout_test.go) |
| Supplied rectangle paint (.20) | Opt-in native software plans receive the actual destination and ancestor clip, including outline outsets. Default painting remains available when the provider is nil; GL rejects supplied plans explicitly. | [API](canvas/rectangle_painter.go), [software controls](internal/painter/software/rectangle_painter_test.go), [real tree and order tests](internal/painter/software/rectangle_painter_integration_test.go), [GL rejection](internal/painter/gl/rectangle_painter_test.go) |
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

Rendering fixes do not establish general CSS support or universal browser pixel parity. The
[Astro Fyne converter](https://github.com/thinkddr/astro-fyne) maintains its own
supported-source contracts and visual comparisons.

For the general toolkit tutorial, packaging and examples, read the preserved
[upstream README](UPSTREAM-README.md) and
[Fyne developer documentation](https://developer.fyne.io/).
Report fork-specific issues in [this repository](https://github.com/thinkddr/fyne/issues).
