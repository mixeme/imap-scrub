# Roadmap

Planned work for `mixeme/imap-scrub`, tracked against [axllent/imap-scrub issues](https://github.com/axllent/imap-scrub/issues).

Shipped changes are in [Changelog](CHANGELOG.md).

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
