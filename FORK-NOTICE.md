# Fork notice

This repository contains a downstream modification of
[Fyne v2.8.1](https://github.com/fyne-io/fyne/tree/v2.8.1). The documented release
is `v2.8.1-sytue.15`, source revision `3afa9294`. It preserves the canonical Go
module path `fyne.io/fyne/v2` and the upstream dependency declarations.

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

## Release delta

These are the 17 commits after upstream `v2.8.1`, in chronological order.
Links identify the exact changes rather than a claim of equivalent behavior on
every platform.

| Commit | Change |
| --- | --- |
| [fa77f33f](https://github.com/thinkddr/fyne/commit/fa77f33f) | Add the optional incoming-URL API, queue and platform bridges. |
| [f9765fe1](https://github.com/thinkddr/fyne/commit/f9765fe1) | Publish fork terms under upstream BSD 3-Clause. |
| [92e078e1](https://github.com/thinkddr/fyne/commit/92e078e1) | Compare cache expiry with the sample used to mark entries alive. |
| [d7a9f8d7](https://github.com/thinkddr/fyne/commit/d7a9f8d7) | Queue background `fyne.Do` calls for execution on the test goroutine. |
| [463ecbc7](https://github.com/thinkddr/fyne/commit/463ecbc7) | Add opt-in dismissing-tap/wheel pass-through and focus scopes. |
| [2116ba8a](https://github.com/thinkddr/fyne/commit/2116ba8a) | Add Android multiple selection and content-provider file size queries. |
| [6ebf66b9](https://github.com/thinkddr/fyne/commit/6ebf66b9) | Intersect nested clipping bounds. |
| [759e3286](https://github.com/thinkddr/fyne/commit/759e3286) | Apply theme line spacing consistently to multiline Entry rows. |
| [2f585d56](https://github.com/thinkddr/fyne/commit/2f585d56) | Round software-painter positions to the nearest device pixel. |
| [8052a020](https://github.com/thinkddr/fyne/commit/8052a020) | Deliver iOS Universal Links through the URL receiver. |
| [537259f2](https://github.com/thinkddr/fyne/commit/537259f2) | Queue mobile dispatch calls made before the event loop starts. |
| [30097551](https://github.com/thinkddr/fyne/commit/30097551) | Adopt the iOS scene lifecycle and its URL callbacks. |
| [85391e40](https://github.com/thinkddr/fyne/commit/85391e40) | Deliver macOS opened documents as `file://` URLs. |
| [098d4660](https://github.com/thinkddr/fyne/commit/098d4660) | Reduce retention from theme overrides and reuse theme-font faces across scopes. |
| [63aefe4d](https://github.com/thinkddr/fyne/commit/63aefe4d) | Use separate GL/ES alpha blending to preserve opaque framebuffer edges. |
| [e5a516fd](https://github.com/thinkddr/fyne/commit/e5a516fd) | Key theme-font faces by style, resource name, backing slice address and length. |
| [3afa9294](https://github.com/thinkddr/fyne/commit/3afa9294) | Map Android key event actions correctly to press/release directions. |

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
