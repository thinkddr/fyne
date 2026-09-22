# Sytue URL-reception patch

This public repository is a fork of [Fyne v2.8.1](https://github.com/fyne-io/fyne)
for the temporary `URLHandler` API and its Android, iOS and macOS bridges.

Since v2.8.1-sytue.3 it also carries a fix in `internal/cache`: `Clean` measured
expiry against the clock sample it had just taken, while `setAlive` stamps entries
with the sample the previous `Clean` left behind, so an entry created after a long
gap without frames (a test that captures nothing for over `ValidDuration`) was
expired at birth and its renderer re-created without `Layout`.

The upstream `LICENSE` and `AUTHORS` are retained unchanged. Original Fyne code
remains copyright its contributors; Sytue's additions are copyright (C) 2026
Sytue. The complete combined work, including these modifications, is distributed
under the upstream BSD 3-Clause License in `LICENSE`.

This fork is not affiliated with or endorsed by Fyne.io. The patch will be
proposed upstream; if it is accepted, consumers will return to upstream Fyne and
this fork will be retired.
