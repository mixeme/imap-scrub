# Roadmap

Planned work for `mixeme/imap-scrub`, tracked against [axllent/imap-scrub issues](https://github.com/axllent/imap-scrub/issues).

Shipped changes are in [Changelog](CHANGELOG.md).

## Recently shipped

Phase 1 & 2 of the [plan](ROADMAP_PLAN.md) are complete (currently in `[Unreleased]`, targeting `0.2.0`):

| Source | Summary | Notes |
| --- | --- | --- |
| [#6](https://github.com/axllent/imap-scrub/issues/6) | Preserve S/MIME signed messages on `remove_attachments` | New per-rule `keep_signatures` (default `true`); S/MIME messages are skipped whole rather than having parts stripped, since removing any part invalidates the signature |
| [#9](https://github.com/axllent/imap-scrub/issues/9) | Multi-folder rules via `mailbox` wildcards | Uses native IMAP `LIST` wildcards (`*` / `%`); one rule expands to every matching folder. (No regex — server-side globbing only) |

## Features

| Source | Summary | Notes |
| --- | --- | --- |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | OAuth login (mainly Gmail) | Password IMAP only today |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | Downscale attachments instead of removing | Smaller in-message copy for large images (incl. inline) |
| — | `export_mailbox` follow-ups | Resume append by Message-ID; separate export path; Homebrew formula |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | Broader rewrite for flexibility | Long term |

### Suggested order

1. `export_mailbox` follow-ups
2. OAuth ([#10](https://github.com/axllent/imap-scrub/issues/10))
3. Attachment downscaling ([#10](https://github.com/axllent/imap-scrub/issues/10))
4. Rewrite ([#10](https://github.com/axllent/imap-scrub/issues/10))
