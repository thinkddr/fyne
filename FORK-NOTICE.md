# Fork notice

This repository contains a downstream modification of
[Fyne v2.8.1](https://github.com/fyne-io/fyne/tree/v2.8.1). The documented release
is [v2.8.1-sytue.16](https://github.com/thinkddr/fyne/tree/v2.8.1-sytue.16).
It preserves the canonical Go module path `fyne.io/fyne/v2` and the upstream
dependency declarations.

## Attribution and license

The upstream [LICENSE](LICENSE) and [AUTHORS](AUTHORS) are retained unchanged.
Original Fyne code remains copyright its contributors. Sytue additions are
copyright © 2026 Sytue and are also distributed under the same **BSD 3-Clause
license**. The same terms cover the combined toolkit code and permit public
redistribution. Bundled fonts retain their separate
[license notices](theme/font/).

Keep the license, copyright notices and disclaimer with redistributions as
required by LICENSE. Fyne.io and its contributors do not endorse this independent
downstream release. General upstream documentation is preserved in
[UPSTREAM-README.md](UPSTREAM-README.md).

## Maintenance release v2.8.1-sytue.16

The [maintenance comparison](https://github.com/thinkddr/fyne/compare/v2.8.1-sytue.15...v2.8.1-sytue.16)
records the complete changes since `.15`:

- English fork documentation and the preserved upstream README.
- Pinned Linux x86-64 regression CI and unchanged upstream reference runs on
  x86-64 and ARM64, with failed-image artifacts retained for review.
- Font-cache backing-storage identity recorded without unsafe operations, and a
  helper extracted from dismissed-overlay dispatch to reduce function complexity.
- Watcher tests that synchronize background callbacks on the test goroutine and
  close resources, with production watcher behavior preserved.
- Entry refresh updates its scroll bounds after text shrinks, clearing stale
  offsets when restored content fits the viewport.
- Reviewed visual fixtures reflecting the existing nearest-pixel position
  rounding and multiline Entry row-spacing changes.

The fixture refresh does not establish browser pixel parity. The platform and
cache boundaries below still apply. [FORK-VALIDATION.md](FORK-VALIDATION.md)
records the reference runs and fixture inventory.

## Initial fork delta v2.8.1-sytue.15

These are the 17 commits after upstream `v2.8.1` through `v2.8.1-sytue.15`
(`d943ebc7`), in chronological order. Links identify the exact changes rather
than equivalent behavior on every platform. The `.15` tag remains unchanged.

| Commit | Change |
| --- | --- |
| [2430cc17](https://github.com/thinkddr/fyne/commit/2430cc17) | Add the optional incoming-URL API, queue and platform bridges. |
| [0e48c701](https://github.com/thinkddr/fyne/commit/0e48c701) | Publish fork terms under upstream BSD 3-Clause. |
| [74da4aea](https://github.com/thinkddr/fyne/commit/74da4aea) | Compare cache expiry with the sample used to mark entries alive. |
| [e337b0f2](https://github.com/thinkddr/fyne/commit/e337b0f2) | Queue background `fyne.Do` calls for execution on the test goroutine. |
| [faf352b9](https://github.com/thinkddr/fyne/commit/faf352b9) | Add opt-in dismissing-tap/wheel pass-through and focus scopes. |
| [6bd31abe](https://github.com/thinkddr/fyne/commit/6bd31abe) | Add Android multiple selection and content-provider file size queries. |
| [cbbc17d7](https://github.com/thinkddr/fyne/commit/cbbc17d7) | Intersect nested clipping bounds. |
| [a353ee54](https://github.com/thinkddr/fyne/commit/a353ee54) | Apply theme line spacing consistently to multiline Entry rows. |
| [4d4ebf89](https://github.com/thinkddr/fyne/commit/4d4ebf89) | Round software-painter positions to the nearest device pixel. |
| [80f85dbc](https://github.com/thinkddr/fyne/commit/80f85dbc) | Deliver iOS Universal Links through the URL receiver. |
| [ae1ae788](https://github.com/thinkddr/fyne/commit/ae1ae788) | Queue mobile dispatch calls made before the event loop starts. |
| [417dce3e](https://github.com/thinkddr/fyne/commit/417dce3e) | Adopt the iOS scene lifecycle and its URL callbacks. |
| [b2a8bdfc](https://github.com/thinkddr/fyne/commit/b2a8bdfc) | Deliver macOS opened documents as `file://` URLs. |
| [77c580ac](https://github.com/thinkddr/fyne/commit/77c580ac) | Reduce retention from theme overrides and reuse theme-font faces across scopes. |
| [5a25290d](https://github.com/thinkddr/fyne/commit/5a25290d) | Use separate GL/ES alpha blending to preserve opaque framebuffer edges. |
| [bec3cd53](https://github.com/thinkddr/fyne/commit/bec3cd53) | Key theme-font faces by style, resource name, backing slice address and length. |
| [d943ebc7](https://github.com/thinkddr/fyne/commit/d943ebc7) | Map Android key event actions correctly to press/release directions. |

## Compatibility boundaries

- `URLHandler` is separate from `fyne.App`, so third-party App implementations
  need not add the method. Scheme/document registration and iOS associated-domain
  configuration belong to the application. Linux/Windows only parse startup URL
  arguments. The bridge does not validate application authentication state.
- `ShowFileOpenMultiple` enables Android multiple selection. iOS and desktop use
  a single-file fallback; callers own every returned reader. Android URI size is
  an optional provider result, not a new `fyne.URI` method.
- Overlay/focus behavior is opt-in through structural methods on application
  objects. A dismissing overlay must permit tap pass-through; wheel pass-through
  is implemented in the GLFW desktop path. Focus traversal uses the first visible
  declared subtree.
- Weak theme keys require Go 1.24 or newer. Go 1.22/1.23 use strong keys. The
  retention regression does not prove that all cache metadata is bounded.
- Theme-font reuse applies to theme-selected faces, not arbitrary custom sources.
  Its key uses backing-storage identity, not a content hash; equal independent
  allocations remain distinct. Explicit custom sources still require comparable
  `fyne.Resource` values. Keep font storage stable while cached.
- Software position rounding and GL alpha changes are targeted fixes, not a CSS
  renderer or a guarantee of pixel equivalence with Chromium.

[README.md](README.md) links the relevant implementations and regression tests.
Native OS callbacks, mobile pickers, iOS lifecycle and GL compositor behavior also
need platform validation; a headless test alone cannot establish those results.
