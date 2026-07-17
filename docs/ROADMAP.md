# Roadmap

Planned work for `mixeme/imap-scrub`, tracked against [axllent/imap-scrub issues](https://github.com/axllent/imap-scrub/issues).

Shipped changes are in [Changelog](CHANGELOG.md).

## Features

| Source | Summary | Notes |
| --- | --- | --- |
| [#6](https://github.com/axllent/imap-scrub/issues/6) | Preserve S/MIME signatures on `remove_attachments` | Keep `smime.p7m`, `smime.p7s`, `smime.p7z` ([RFC 8551](https://www.rfc-editor.org/rfc/rfc8551)); optional config, default on |
| [#9](https://github.com/axllent/imap-scrub/issues/9) | Regex or glob `mailbox`; apply one rule to many folders | e.g. `mailbox: "INBOX.*"`. Workaround: `-m` → YAML list → generate rules externally ([comment](https://github.com/axllent/imap-scrub/issues/9#issuecomment-3167450004)) |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | OAuth login (mainly Gmail) | Password IMAP only today |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | Downscale attachments instead of removing | Smaller in-message copy for large images (incl. inline) |
| — | `export_mailbox` follow-ups | Resume append by Message-ID; separate export path; Homebrew formula |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | Broader rewrite for flexibility | Long term |

### Suggested order

1. S/MIME ([#6](https://github.com/axllent/imap-scrub/issues/6))
2. Regex / multi-mailbox ([#9](https://github.com/axllent/imap-scrub/issues/9))
3. `export_mailbox` follow-ups
4. OAuth ([#10](https://github.com/axllent/imap-scrub/issues/10))
5. Attachment downscaling ([#10](https://github.com/axllent/imap-scrub/issues/10))
6. Rewrite ([#10](https://github.com/axllent/imap-scrub/issues/10))
