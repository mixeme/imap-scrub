# Roadmap

Planned work for `mixeme/imap-scrub`, tracked against [axllent/imap-scrub issues](https://github.com/axllent/imap-scrub/issues).

Shipped changes are in [Changelog](CHANGELOG.md).

## Features

| Source | Summary | Notes |
| --- | --- | --- |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | Downscale attachments instead of removing | Smaller in-message copy for large images (incl. inline) |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | Broader rewrite for flexibility | Long term |

OAuth login ([#10](https://github.com/axllent/imap-scrub/issues/10), Gmail XOAUTH2) and the `export_mailbox` follow-ups (resume append by Message-ID, separate `export_path`, Homebrew formula) shipped — see [Changelog](CHANGELOG.md).

### Suggested order

1. Attachment downscaling ([#10](https://github.com/axllent/imap-scrub/issues/10))
2. Rewrite ([#10](https://github.com/axllent/imap-scrub/issues/10))
