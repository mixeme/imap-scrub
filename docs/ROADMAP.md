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

## Dependencies

### Migrate away from archived `gopkg.in/yaml.v3`

[go-yaml/yaml](https://github.com/go-yaml/yaml) was archived in April 2025. We use `gopkg.in/yaml.v3 v3.0.1` (`lib/config.go`, `lib/imap_integration_test.go`).

- Short term: [`go.yaml.in/yaml/v3`](https://github.com/yaml/go-yaml)
- Longer term: [`go.yaml.in/yaml/v4`](https://github.com/yaml/go-yaml)

Import path change only; `Marshal` / `Unmarshal` stay compatible.

### Review other dependencies

| Module | Current | Notes |
| --- | --- | --- |
| `gopkg.in/yaml.v3` | v3.0.1 | Archived — migrate (see above) |
| `github.com/apsdehal/go-logger` | v0.0.0-201905… | Unmaintained since 2019 — consider `log/slog` |
| `github.com/axllent/semver` | v0.0.1 | v1.0.0 available |
| `github.com/emersion/go-imap` | v1.2.1 | Up to date |
| `github.com/emersion/go-message` | v0.18.1 | v0.18.2 available |
| `github.com/emersion/go-mbox` | v1.0.2 | Up to date |
| `github.com/spf13/pflag` | v1.0.5 | v1.0.10 available |
| `golang.org/x/text` (indirect) | v0.14.0 | v0.40.0 available |

**Suggested order:** YAML → patch bumps (`pflag`, `go-message`, `semver`) → replace `go-logger` → revisit Go minimum version if needed.
