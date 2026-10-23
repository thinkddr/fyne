# Fork notice

This repository contains a downstream modification of
[Fyne v2.8.2](https://github.com/fyne-io/fyne/tree/v2.8.2). This document describes
source version `v2.8.2-sytue.1`. Published versions are listed in
[Tags](https://github.com/thinkddr/fyne/tags); the prior release is
[v2.8.1-sytue.21](https://github.com/thinkddr/fyne/tree/v2.8.1-sytue.21).
It preserves the canonical Go module path `fyne.io/fyne/v2` and the upstream
dependency declarations.

## Attribution and license

The upstream [LICENSE](LICENSE) and [AUTHORS](AUTHORS) are retained unchanged.
Original Fyne code remains copyright its contributors under the **BSD 3-Clause
license**. The changes and additions made in this fork are copyright © 2026
Daniel Delgado Ramos and are licensed under the **Apache License 2.0**
([LICENSE-MODIFICATIONS](LICENSE-MODIFICATIONS)). Both licenses permit public
redistribution. Pull requests are welcome; @thinkddr is the sole
maintainer ([CONTRIBUTING.md](CONTRIBUTING.md)). Bundled fonts retain their separate
[license notices](theme/font/).

Keep the license, copyright notices and disclaimer with redistributions as
required by LICENSE. Fyne.io and its contributors do not endorse this independent
downstream release. General upstream documentation is preserved in
[UPSTREAM-README.md](UPSTREAM-README.md).

## Upstream 2.8.2 integration in v2.8.2-sytue.1

This source integrates the official stable upstream `v2.8.2` commit
`77c1fa8e0a50b6cbd04cd6c4454289e55a307a2d` with fork parent
`bf85664117f0a3acc2cf3388e07fd7dee32a3d38` (`v2.8.1-sytue.21`). Neither published
tag is changed. The retained fork history and the current upstream release history
have different commit IDs for their 2.8.1 base. Fork commit
`1138eac0a9ccaa7daf1e9b9c7236bd32b9abd188` and official upstream 2.8.1 commit
`3dc06f47137aa3709807f1bb78dacc5daf68d551` have the identical Git tree
`ba72c7168235d955109a1257c02afe64d93fe2a9`. The merge uses that explicit content
base and records both release parents; it does not rewrite either history.

The integration preserves incoming URLs, picker behavior, touch arbitration,
cancellation and all supplied-painter hooks. iOS retains `urlOpened` alongside the
new demand-driven render loop. Aligned text selection combines upstream's row
alignment with the fork's row spacing. The fork's coverage threshold remains 62;
upstream's lower threshold is not adopted without a measured need.

The Android picker applies upstream's persistent-read grant to every selected URI,
including multi-selection. It requests only read access, retains it only when the
provider's result grants persistence, and preserves temporary access if that
provider refuses persistence. It does not log the selected URI.

Go remains 1.22. The canonical module list is unchanged; upstream updates
`github.com/go-text/typesetting` to `v0.3.5`, `github.com/go-gl/glfw/v3.4/glfw` to
`v0.1.0-pre.2`, and `github.com/fyne-io/oksvg` to
`v0.2.1-0.20260918172519-a9af55fa95d9`. Consumers with replacements for those
modules must validate their actual replacement source separately. Passing this
fork's tests against canonical dependencies does not certify older replacements.

## Supplied rectangle painting in v2.8.1-sytue.20

Version `.20` adds opt-in [`canvas.Rectangle.Painter`](canvas/rectangle_painter.go)
through `RectanglePainter`. A provider supplies an immutable software paint plan
and owns its resources. It receives the actual accumulated NRGBA destination,
original stride, unrounded global logical origin, logical rectangle size, canvas
scale and physical ancestor clip. The destination also intersects canvas bounds;
it allows outline outsets beyond the rectangle's own box.

A supplied plan owns the complete rectangle paint, including any fill, stroke,
corner, aspect and shadow effects. Providers must respect the clip and validate
unsupported requests before modifying pixels; Fyne cannot roll back partial
writes. Invalid non-finite or negative geometry fails before calling the provider.
`PaintError` retains failures across measurement, refresh and fully clipped
paints. A successful supplied paint or explicit `SetPaintError(nil)` clears the
error. GL reports `ErrRectanglePainterUnsupportedGL` without calling the software
provider or drawing a default rectangle. A nil painter preserves the existing
rectangle path. No CGo or native-engine dependency is added.

Unit and real-canvas regressions cover the accumulated backdrop, original stride,
global fractional position, ancestor clipping and outsets, and sibling paint order
at scales 1 and 2. Separate controls exercise invalid geometry, sticky errors,
recovery and the default path. The GL regression checks explicit refusal.
These hooks do not implement CSS outlines or establish browser, native-widget,
editor, inverse-conversion or GPU
parity. Consumers must reject or explicitly represent custom paint plans when
exporting a scene.

The implementation at `a82d070eff48adf2845f7887db7c82bb0d4488d2` passed all five
fork workflows, with thirteen successful jobs:

- [Fork conformance](https://github.com/thinkddr/fyne/actions/runs/38962828156):
  canonical Linux regressions; the tag-only `public-install` job was skipped.
- [Platform Tests](https://github.com/thinkddr/fyne/actions/runs/38962831278):
  all six Linux, macOS and Windows jobs, including Linux Wayland and coverage.
- [Static Analysis](https://github.com/thinkddr/fyne/actions/runs/38962836187):
  all three jobs.
- [Mobile Tests](https://github.com/thinkddr/fyne/actions/runs/38962840493):
  declared and stable Go.
- [Web Tests](https://github.com/thinkddr/fyne/actions/runs/38962844104).

These branch runs exclude a public-module install. Release validation reruns all
five workflows on the final source and tag; the tag-only clean install must
select the actual published `.20` module and compile the new rectangle API along
with the existing URL and dialog APIs. Consult
[Actions](https://github.com/thinkddr/fyne/actions) for those results. Earlier
release or consumer pixel evidence does not certify this hook or `.20` pixels.

## Software shaped-text in v2.8.1-sytue.19

Version `.19` adds an opt-in [`canvas.Text.Layout`](canvas/text_layout.go)
interface. It supplies immutable line-box metrics and glyph painting into
the software canvas's accumulated, clipped NRGBA buffer. The provider owns font
resources and must respect the actual stride, physical clip, logical origin,
scale and resolved color. Unsupported requests must fail before pixel changes;
Fyne cannot roll back provider writes. No CGo or font-engine dependency is added.

Metric failures remain visible and never trigger default-font measurement.
Paint failures remain visible across metric queries. Successful software paint
clears its paint error; an explicit setter can also clear it. GL rejects supplied
layouts with `ErrTextLayoutUnsupportedGL`, with no default glyph or opaque tile
substitution. The nil layout path remains unchanged.

Fyne paints underline and strikethrough after successful glyph paint, using the
supplied width and baseline with its existing stroke and offsets. Hook-only
decoration blits account for positive clip offsets. Unsafe coordinate/allocation
requests are rejected before provider painting, including decoration rasters
larger than 16 Mi pixels. This does not implement arbitrary CSS decorations or
establish native-widget, browser, editor, inverse-conversion or GPU parity.
Regressions exercise metrics, errors, destination/backdrop access,
clipping, decorations and the default path.

The implementation at `eb30b4793d6eb857ae080acaa24d80a4c4ea6c19` passed all five
fork workflows, with thirteen successful jobs:

- [Fork conformance](https://github.com/thinkddr/fyne/actions/runs/38930738931):
  canonical Linux regressions; the tag-only `public-install` job was skipped.
- [Platform Tests](https://github.com/thinkddr/fyne/actions/runs/38930748038):
  all six Linux, macOS and Windows jobs.
- [Static Analysis](https://github.com/thinkddr/fyne/actions/runs/38930756806):
  all three jobs.
- [Mobile Tests](https://github.com/thinkddr/fyne/actions/runs/38930766117):
  declared and stable Go.
- [Web Tests](https://github.com/thinkddr/fyne/actions/runs/38930774376).

These runs apply to that source commit and do not include a public-install
result. Release validation reruns all five workflows for the final source and
tag; the clean public-module install must resolve the actual published `.19`
module. Consult [Actions](https://github.com/thinkddr/fyne/actions) for the
corresponding results. Consumers must separately validate any supplied font
engine and their rendering profile; the `.17` image evidence below does not
certify this text hook or `.19` pixels.

## Maintenance release v2.8.1-sytue.18

The [`.17` to `.18` comparison](https://github.com/thinkddr/fyne/compare/v2.8.1-sytue.17...v2.8.1-sytue.18)
records the cache expiry fix and its regression. Each expiry timestamp now uses
`atomic.Int64`, with matching loads for expiry checks and stores for refreshes.
The cached clock sample, duration arithmetic and strict `<` comparison are
unchanged. An entry remains alive exactly at its expiration timestamp.

[Fresh `.17` main CI](https://github.com/thinkddr/fyne/actions/runs/38899942135/job/116969412331)
reported concurrent expiry reads during renderer cleanup and refresh writes
during a menu markup assertion. The same-tree `.17` tag jobs passed. The cache
source involved was unchanged from the `.16` base; neither that provenance nor a
passing tag run dismisses the observed race.

The new concurrent regression was first tested against the original expiry
implementation in [run 38901005903](https://github.com/thinkddr/fyne/actions/runs/38901005903).
It failed under the race detector with both read/write and write/write reports
in `TestExpiringCacheConcurrentRefresh`. The test also checks that entries stay
alive immediately before and exactly at expiry, then expire immediately after.
No test suppression, image fixture update or tolerance change accompanies the fix.

The `.18` final candidate must pass the full fork workflows before publication.
Clean public installation and strict image comparisons against the actual `.18`
module remain separate release checks. The `.17` pixel evidence below does not
transfer a certification to `.18`.

## Software image release v2.8.1-sytue.17

The [`.16` to `.17` comparison](https://github.com/thinkddr/fyne/compare/v2.8.1-sytue.16...v2.8.1-sytue.17)
records this release's changes. The sampler preserves the existing image
decode/cache/pre-scaling lifecycle,
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

The checked RGBA8 conversion rejects an internal out-of-range result instead of
silently narrowing it. Valid byte inputs and Q4 weights retain the same sampling
arithmetic. Arithmetic and invalid-weight regressions cover that boundary.

### Release checks and evidence

[Converter run 38893097982](https://github.com/thinkddr/astro-fyne/actions/runs/38893097982)
at `27168b4c8a21a458542904e345ee86f9fcb8b9e7` tested an earlier sampling candidate.
It passed all six strict RGBA frame
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

That run predates the final checked-conversion guard and formatting changes; its
pixel results do not certify the final `.17` package. The actual published
`.17` module was then tested in
[run 38900188638](https://github.com/thinkddr/astro-fyne/actions/runs/38900188638).
All six strict RGBA comparisons passed with source, license and lock-file
stability checks. That evidence belongs to the published `.17` module and the
recorded profile, not another release.

The `.17` tag at `46ef36fda23a9e5138bf422505b2ce760746cc20` passed all five
complete fork workflows:

- [Platform Tests](https://github.com/thinkddr/fyne/actions/runs/38899942393):
  declared and stable Go on Linux, macOS and Windows, including Linux Wayland
  checks and the required coverage check.
- [Static Analysis](https://github.com/thinkddr/fyne/actions/runs/38899942498):
  all three jobs, including lint and formatting.
- [Mobile Tests](https://github.com/thinkddr/fyne/actions/runs/38899942427):
  declared and stable Go.
- [Web Tests](https://github.com/thinkddr/fyne/actions/runs/38899942415).
- [Fork conformance](https://github.com/thinkddr/fyne/actions/runs/38899942413):
  canonical Linux regressions and a clean public-module install using canonical
  Fyne imports. The install resolved `github.com/thinkddr/fyne/v2` at `.17`,
  printed module verification success and built successfully.

The same-tree main stable Linux job exposed the cache race described in the
`.18` maintenance section above. All failures remain part of the release record.

This evidence is limited to the recorded software-renderer profile. It does not
certify other physical factors, fractional placement, arbitrary backgrounds,
color profiles, rounded clips, typography or GPU output.

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

## Optional mobile touch recognition in v2.8.1-sytue.21

`driver/mobile.TouchFilterCanvas` exposes an opt-in canvas callback for resolving
single-finger navigation and refresh gestures before dispatch reaches children.
It buffers an undecided sequence, replays rejected events in order with their
original times, and consumes claimed sequences. A second finger cancels pending
recognition and preserves ordinary multi-touch; after a claim it cancels the
recognizer and consumes remaining fingers. Root, overlay, geometry and filter changes cancel pending
recognition. A visible, clipped hit path is available at begin and end so callers
can verify that their view still owns the gesture. No new method is added to the
canonical `fyne.Canvas` interface and desktop behavior is unchanged.

The native bridges distinguish cancellation from release. iOS forwards the real
`touchesCancelled:withEvent:` callback and keeps still-active finger IDs stable;
Android forwards `ACTION_CANCEL` for every active pointer. Cancellation does not
produce a tap or fling. This addition requires native device validation in
addition to the driver regression tests; a headless pass alone does not validate
OS gesture arbitration.

This release includes the earlier rectangle-paint source documented as `.20`.
It does not publish or move that reserved tag. The mobile callback adds no Android
system-gesture exclusion; consumers must arbitrate any such bounded region with
the platform separately and preserve the system's mandatory gestures.
