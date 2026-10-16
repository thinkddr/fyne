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

## Unreleased software image enlargement

The candidate preserves the existing image decode/cache/pre-scaling lifecycle,
identity-size shortcut, clipping and final compositing. Only software
`canvas.Image` calls using `ImageScaleSmooth`, object alpha exactly 1 and corner
radius exactly 0 can use the new sampler. `canvas.Raster`, Pixels and Fastest keep
their existing routes. A declined Smooth Image profile retains the original
CatmullRom fallback; there is no broad bilinear replacement.

The helper checks the buffer returned by `painter.PaintImage` and its
`scaleImage` prepass: NRGBA8 or opaque RGBA8, zero bounds origin, positive
dimensions, uniform physical factor exactly 2 or 4, destination axes at most
16,384 and area at most 64 Mi pixels. The prepass may normalize original 16-bit,
nonzero-origin or translucent RGBA sources into eligible NRGBA buffers. These
checks do not guarantee that those original sources retain their old sampling.
Direct declined-profile tests exercise the post-prepass boundary.

The interpolation derives from Skia revision
`9d07e5bad9e3e21da2426946e589daa647218271`. It accumulates four Q4-by-Q4 weights
before flooring interpolated premultiplied RGB and alpha separately. The
[sampler](internal/painter/software/skia_legacy_enlargement.go) retains its Google
LLC and Fyne contributors notices. Keep [LICENSE-Skia.txt](LICENSE-Skia.txt) and
[LICENSE-Fyne.txt](LICENSE-Fyne.txt) with this code; existing toolkit and bundled
font notices remain in force.

[Converter run 38893097982](https://github.com/thinkddr/astro-fyne/actions/runs/38893097982)
at `27168b4c8a21a458542904e345ee86f9fcb8b9e7` passed all six strict RGBA frame
comparisons for three original 16-by-16 PNG8 assets displayed at 32-by-32 CSS
pixels, over opaque white, at DPR 1 and 2. Assets include odd RGB values, an
explicit 1-to-3 rounding control and alpha boundaries. The comparisons cover
source-only generated native layout, native-to-web export and reconstructed
native scenes; exported resources retain the original PNG bytes.

The complete `internal/painter/software`, `internal/painter` and `canvas`
packages passed with the race detector on canonical Linux x86-64 under the
converter-resolved probe graph and fork-own graph: x/image v0.46.0 and v0.24.0
respectively. The probe is not a claim of complete native module-graph identity.
No image fixture was updated. Fixed arithmetic vectors, opt-in/declined-profile
comparisons and a real Image/Raster routing test remain independent controls.

This evidence is limited to the recorded software-renderer profile. It does not
certify other physical factors, fractional placement, arbitrary backgrounds,
color profiles, rounded clips, typography or GPU output. Full fork release checks
on the exact candidate remain required before publishing a new version.

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
- Menu tests exercise later mouse and keyboard actions even if an initial image
  assertion fails, preserving all image, markup and callback assertions.
- Coverage records repository code executed across package integration tests,
  retains the 62% requirement and archives the report. Coveralls upload is opt-in.

The fixture refresh does not establish browser pixel parity. The platform and
cache boundaries below still apply. [FORK-VALIDATION.md](FORK-VALIDATION.md)
records the reference runs and fixture inventory.

## Initial fork delta v2.8.1-sytue.15

These are the 17 commits after upstream `v2.8.1` through `v2.8.1-sytue.15`
(`3afa9294`), in chronological order. Links identify the exact changes rather
than equivalent behavior on every platform. The `.15` tag remains unchanged.

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
