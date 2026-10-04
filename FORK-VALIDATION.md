# Fork validation

The software-image reference uses Linux x86-64, Go 1.27.1 and the repository's
locked dependencies. It checks the fork's behavior, not browser pixel parity or
native OS integrations.

## Reference and fixture review

The [unchanged upstream reference](https://github.com/thinkddr/fyne/actions/runs/37198815282)
at `3dc06f47137aa3709807f1bb78dacc5daf68d551` passes all tests on Linux x86-64.
Linux ARM64 has image differences even with that unchanged source and toolchain.
The strict fork reference therefore uses x86-64 without changing pixel tolerance.

The [rendering diagnostic](https://github.com/thinkddr/fyne/actions/runs/37199895570)
tested the fork with exactly three existing changes reversed: nearest-pixel
software positions, multiline Entry row spacing and nested clip intersection.
It restored the 55 masters previously changed by those fixes to upstream bytes.

Both desktop and mobile variants completed all 67 packages with zero legacy
image mismatches, no races and no unexpected failures. Exactly these three
regression controls failed, as required by the diagnostic:

- `TestPainter_fractionalPositionRounds`
- `TestEntry_MultiLineRowsFollowLineSpacing`
- `TestWalkVisibleObjectTree_NestedClipIntersects`

That result identifies the stale fixtures as consequences of the intended
rendering changes. The temporary diagnostic is retained in Git history and its
CI evidence; production tests use the real fork implementation.

The maintenance release refreshes 116 PNG masters: 108 desktop fixtures and eight
additional mobile fixtures. All 83 shared desktop/mobile captures match exactly
in RGBA. Dimensions remain unchanged. The
[fixture inventory](ci/fork-visual-reference.json) records source revisions,
diagnostic results and every before/after SHA-256 hash.

Twelve menu captures were exposed after correcting an earlier initial image:
the original test skipped its action steps when that assertion failed. Those
steps all passed in the rendering diagnostic. The test now always exercises
every mouse and keyboard action, while retaining each image, markup and callback
assertion.

The fixture review also exposed stale Entry scroll geometry after undo. The
[viewport regression](widget/entry_scroll_internal_test.go) verifies that typing
an overflowing row and then undoing it restores the text, cursor, fitting bounds,
zero scroll offset and the exact initial capture. Entry refresh now updates the
active scroll layout before cursor visibility is calculated.
The initial and five-undo image masters use the corrected fitting viewport;
their inventory entries retain the earlier capture hashes and correction notes.

## Release checks

[Fork conformance](.github/workflows/fork-conformance.yml) runs module
verification, formatting, vet and race-enabled tests against the real source.
The inherited platform, mobile, web and static-analysis workflows remain active.
Linux platform tests retain the upstream 62% coverage requirement and archive
the coverage report. Coverage now instruments every repository package in every
test with `-coverpkg=./...`, so software-renderer and widget integration tests
record the library code they exercise across package boundaries. This is a
module-wide execution metric; Go's default only measures the package under
test. See the [Go test coverage flags](https://pkg.go.dev/cmd/go#hdr-Testing_flags).
External Coveralls uploads require the repository variable
`COVERALLS_ENABLED=true` and enrollment of this fork with that service.
Release tags also build a clean external module using canonical Fyne imports and
the published fork replacement from the README.

Platform tests and software captures do not validate native URL callbacks,
Android/iOS pickers, iOS scenes or GL compositor output. Those require platform
or device testing. Rendering changes do not establish universal CSS support or
Chromium glyph equivalence.
