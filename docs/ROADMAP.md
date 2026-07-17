# Roadmap

Tracked against [axllent/imap-scrub issues and PRs](https://github.com/axllent/imap-scrub/issues). This fork (`mixeme/imap-scrub`) continues development while upstream is being archived.

Ports from community forks **without** an axllent ticket (jackr0, mikulcak, touste, ahbk, …) are listed only in [Changelog](../CHANGELOG.md).

## Features from upstream

### Open requests

Unimplemented feature ideas still open (or wish-listed) at axllent:

| Source | Summary | Notes |
| --- | --- | --- |
| [#9](https://github.com/axllent/imap-scrub/issues/9) | Regex or glob `mailbox` names; apply one rule to many folders | IMAP lists subfolders as flat names (`INBOX.abc`, …). Suggested: `mailbox: "INBOX.*"` or auto-expand children. Workaround today: `-m` prints YAML list → generate rules externally (see [comment](https://github.com/axllent/imap-scrub/issues/9#issuecomment-3167450004)). |
| [#6](https://github.com/axllent/imap-scrub/issues/6) | Preserve S/MIME digital signatures on `remove_attachments` | Do not strip attachments named `smime.p7m`, `smime.p7s`, or `smime.p7z` ([RFC 8551](https://www.rfc-editor.org/rfc/rfc8551)). Optional config, default keep signatures. |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | OAuth login (mainly Gmail) | README: password IMAP only today. Needed for modern Gmail without app passwords. |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | Downscale attachments instead of removing | Resize large images (including inline) to a smaller copy in the message rather than deleting them entirely. |

### Future improvements (`export_mailbox`)

Core `export_mailbox` is **done** (ported from [axllent#2](https://github.com/axllent/imap-scrub/pull/2), merged for v0.1.0). Optional extras from the original PR author's idea list — **not** open axllent tickets:

- Resume export: append to existing mbox when Message-ID is new
- Separate export directory (not shared with `save_path`)
- Homebrew formula

### Suggested feature order

1. S/MIME signature preservation ([#6](https://github.com/axllent/imap-scrub/issues/6))
2. Regex / multi-mailbox rules ([#9](https://github.com/axllent/imap-scrub/issues/9))
3. OAuth for Gmail ([#10](https://github.com/axllent/imap-scrub/issues/10))
4. Attachment downscaling ([#10](https://github.com/axllent/imap-scrub/issues/10))
5. Broader rewrite for flexibility ([#10](https://github.com/axllent/imap-scrub/issues/10)) — long term

### Implemented (axllent)

| Source | Summary | Status |
| --- | --- | --- |
| [#2](https://github.com/axllent/imap-scrub/pull/2) | `export_mailbox` action → mbox backup | Merged for v0.1.0 — [Changelog](../CHANGELOG.md) |
| [#14](https://github.com/axllent/imap-scrub/pull/14) | `older_than` uses IMAP internal date; local build scripts | Merged for v0.1.0 — [Changelog](../CHANGELOG.md) |
| [#13](https://github.com/axllent/imap-scrub/pull/13) | `pass_file`, comma-separated `from`, `newer_than` | Merged for v0.1.0 — [Changelog](../CHANGELOG.md) |
| [#8](https://github.com/axllent/imap-scrub/pull/8) | Log saved attachment paths to `%d-attachments-deleted.txt` | In upstream 0.0.5 (base of this fork) |
| [#7](https://github.com/axllent/imap-scrub/issues/7) / [#11](https://github.com/axllent/imap-scrub/pull/11) | Skip `remove_attachments` when message has no attachments | In upstream 0.0.6 (base of this fork) |
| [#5](https://github.com/axllent/imap-scrub/issues/5) | Clarify `include_unread` behaviour | In upstream 0.0.4 |
| [#1](https://github.com/axllent/imap-scrub/pull/1) | Fix hang on `Select()` when mailbox list was not consumed | In upstream; `lib/mailboxes.go` |

### Closed upstream (no further work)

| Source | Summary | Notes |
| --- | --- | --- |
| [#12](https://github.com/axllent/imap-scrub/issues/12) | Replace removed attachment with stub text file | Author closed: existing `remove_attachments` behaviour (incl. `%d-attachments-deleted.txt` from [#8](https://github.com/axllent/imap-scrub/pull/8)) is sufficient |
| [#3](https://github.com/axllent/imap-scrub/issues/3), [#4](https://github.com/axllent/imap-scrub/issues/4) | Gmail credentials / nil pointer panic | Closed support/bug reports — not feature requests |

### Maintenance

| Source | Summary | Notes |
| --- | --- | --- |
| [#10](https://github.com/axllent/imap-scrub/issues/10) | Project continuation | Handoff to `mixeme/imap-scrub` agreed; upstream README points here; archive pending |

## Dependencies

### Migrate away from archived `gopkg.in/yaml.v3`

[go-yaml/yaml](https://github.com/go-yaml/yaml) was archived in April 2025 and is no longer maintained. The project currently uses `gopkg.in/yaml.v3 v3.0.1` for config parsing (`lib/config.go`, `lib/imap_integration_test.go`).

**Recommended path:**

- Short term: migrate to [`go.yaml.in/yaml/v3`](https://github.com/yaml/go-yaml) (maintained by the YAML organization; v3 receives security fixes only)
- Longer term: evaluate [`go.yaml.in/yaml/v4`](https://github.com/yaml/go-yaml) when upgrading other tooling (active development)

**Impact:** import path change; `Marshal` / `Unmarshal` API stays compatible for config loading.

### Review other dependencies

| Module | Current | Notes |
| --- | --- | --- |
| `gopkg.in/yaml.v3` | v3.0.1 | Archived upstream — migrate (see above) |
| `github.com/apsdehal/go-logger` | v0.0.0-201905… | Last release 2019; no newer version — consider `log/slog` or another maintained logger |
| `github.com/axllent/semver` | v0.0.1 | v1.0.0 available |
| `github.com/emersion/go-imap` | v1.2.1 | Up to date |
| `github.com/emersion/go-message` | v0.18.1 | v0.18.2 available |
| `github.com/emersion/go-mbox` | v1.0.2 | Added for `export_mailbox`; up to date |
| `github.com/spf13/pflag` | v1.0.5 | v1.0.10 available |
| `golang.org/x/text` (indirect) | v0.14.0 | v0.40.0 available via `go-imap` dependency chain |

**Suggested order:**

1. YAML migration (archived upstream)
2. Patch/minor bumps: `pflag`, `go-message`, `axllent/semver`
3. Assess replacing `go-logger`
4. Revisit Go minimum version if needed for transitive updates
