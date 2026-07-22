# Roadmap

Planned work for `mixeme/imap-scrub`, tracked against [axllent/imap-scrub issues](https://github.com/axllent/imap-scrub/issues).

Shipped changes are in [Changelog](CHANGELOG.md).

## Features

| Source | Summary | Notes |
| --- | --- | --- |
| [#9](https://github.com/axllent/imap-scrub/issues/9) | Regex or glob `mailbox`; apply one rule to many folders | e.g. `mailbox: "INBOX.*"`. Workaround: `-m` → YAML list → generate rules externally ([comment](https://github.com/axllent/imap-scrub/issues/9#issuecomment-3167450004)) |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | OAuth login (mainly Gmail) | Password IMAP only today |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | Downscale attachments instead of removing | Smaller in-message copy for large images (incl. inline) |
| — | `export_mailbox` follow-ups | Resume append by Message-ID; separate export path; Homebrew formula |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | Broader rewrite for flexibility | Long term |

### Suggested order

1. Regex / multi-mailbox ([#9](https://github.com/axllent/imap-scrub/issues/9))
2. `export_mailbox` follow-ups
3. OAuth ([#10](https://github.com/axllent/imap-scrub/issues/10))
4. Attachment downscaling ([#10](https://github.com/axllent/imap-scrub/issues/10))
5. Rewrite ([#10](https://github.com/axllent/imap-scrub/issues/10))
