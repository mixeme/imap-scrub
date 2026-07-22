# Changelog

## [Unreleased]

- **OAuth2 login** ([#10](https://github.com/axllent/imap-scrub/issues/10), see [docs/ROADMAP_PLAN.md](ROADMAP_PLAN.md#1-oauth-login)): new `auth: oauth2` config option authenticates with `AUTHENTICATE XOAUTH2` instead of a password, for accounts where App Passwords are unavailable. Configured via `oauth_client_id`, `oauth_client_secret` and `oauth_token_file`; `pass` / `pass_file` are not required. Defaults target Gmail, and `oauth_auth_url` / `oauth_token_url` / `oauth_scope` override the endpoints for other providers.
  - Authorization is a one-off `--oauth-setup` step. Normal runs only read the token file and refresh the access token, so **no browser is needed on the machine running IMAP-Scrub** — cron, servers and CI work unchanged. A rotated refresh token is written back to disk automatically.
  - `--oauth-setup` opens a browser locally and catches the redirect on a loopback port. `--oauth-headless` (or `OAUTH_HEADLESS=1`) instead prints the URL to open on another device and accepts the code — or the full redirect URL — pasted back on stdin. Token files are written mode `0600`.
- `-p` / `--print-config` now redacts `oauth_client_secret`, and no longer prints a masked `pass` when no password is configured.

### Tests

- OAuth2 coverage runs in the existing unit job — no new CI secrets or network access. Token refresh and refresh-token rotation are tested against an in-process `httptest` endpoint, and the XOAUTH2 handshake against a real in-process IMAP server (`go-imap`'s `server` package + memory backend) so the wire format is verified over an actual `AUTHENTICATE` exchange rather than only asserted as a string.
- The interactive and headless `--oauth-setup` flows are not automated; they were verified manually against a stub provider.

### Dependencies

- Add `golang.org/x/oauth2` (pinned to `v0.30.0`; `v0.36.0` raises the go directive to 1.25, and the project targets 1.23). Google's endpoints are hardcoded as defaults rather than importing `oauth2/google`, which would pull in `cloud.google.com/go/compute/metadata`.
- `github.com/emersion/go-sasl` promoted from an indirect to a direct dependency (used by the handshake tests); no new modules.
- Nix flake `vendorHash` reset to `lib.fakeHash` — run `nix build` once and paste the reported `got:` hash

## [0.3.0] — S/MIME preservation & multi-mailbox rules

- `export_mailbox` follow-ups (see [docs/ROADMAP_PLAN.md](ROADMAP_PLAN.md#3-export_mailbox-follow-ups)):
  - Reruns now **resume** instead of failing: if the mbox file already exists, matching messages are appended to it and any message whose `Message-Id` is already in the file is skipped, so a rule can be safely repeated (e.g. from cron) without duplicating exports. Dry run (no `-y`) reports would-export / would-skip counts by checking the existing mbox read-only.
  - New `export_path` config option writes `mbox` files to a directory separate from `save_path` (used for attachments); falls back to `save_path` when unset.
  - Added a Homebrew formula (`Formula/imap-scrub.rb`) — `brew tap mixeme/imap-scrub https://github.com/mixeme/imap-scrub && brew install imap-scrub`. Stable installs use the prebuilt binaries already attached to each GitHub release (no Go toolchain needed); `--HEAD` builds `develop` from source.
- `mailbox` now accepts IMAP `*` / `%` wildcard patterns ([#9](https://github.com/axllent/imap-scrub/issues/9)), so one rule can apply to many folders (e.g. `mailbox: "INBOX.*"`). Plain mailbox names behave exactly as before.
- `remove_attachments` now preserves S/MIME signed messages by default ([#6](https://github.com/axllent/imap-scrub/issues/6)): skips `application/pkcs7-mime` (`smime.p7m` / `smime.p7z`) and `multipart/signed` (`smime.p7s`) messages instead of stripping their signature parts. New per-rule `keep_signatures` option (default `true`) restores the previous behaviour when set to `false`.

## [0.1.1] — dependency refresh

### Dependencies

- Migrate YAML from archived `gopkg.in/yaml.v3` to maintained `go.yaml.in/yaml/v3`
- Replace unmaintained `apsdehal/go-logger` with `log/slog` (message-only CLI output + ANSI level colors preserved)
- Replace `axllent/semver` with `golang.org/x/mod/semver` (self-updater only)
- Bump `go-message` to v0.18.2, `go-mbox` to v1.0.4, `pflag` to v1.0.10, `golang.org/x/text` to v0.28.0
- Raise minimum Go version to 1.23 (devcontainer, CI, Docker build image)
- Nix flake `vendorHash` reset to `lib.fakeHash` — run `nix build` once and paste the reported `got:` hash

## [0.1.0] — community continuation release

Actively maintained continuation of [axllent/imap-scrub](https://github.com/axllent/imap-scrub) (last upstream release `0.0.6`, Apr 2024).

Install: `go install github.com/mixeme/imap-scrub@latest`  
Releases: https://github.com/mixeme/imap-scrub/releases

### Ports and contributions

| PR | Change | Source |
|----|--------|--------|
| [#1](https://github.com/mixeme/imap-scrub/pull/1) | `older_than` uses IMAP internal date (not envelope `Date`) | [axllent#14](https://github.com/axllent/imap-scrub/pull/14) |
| [#2](https://github.com/mixeme/imap-scrub/pull/2) | Local Windows/Linux build scripts | [axllent#14](https://github.com/axllent/imap-scrub/pull/14) |
| [#3](https://github.com/mixeme/imap-scrub/pull/3) | `pass_file` config option | [axllent#13](https://github.com/axllent/imap-scrub/pull/13) (@Carsten-Leue) |
| [#4](https://github.com/mixeme/imap-scrub/pull/4) | Comma-separated `from` (OR) + unit/integration CI | [axllent#13](https://github.com/axllent/imap-scrub/pull/13) (@Carsten-Leue) |
| [#5](https://github.com/mixeme/imap-scrub/pull/5) | `newer_than` via IMAP internal date | [axllent#13](https://github.com/axllent/imap-scrub/pull/13) (@Carsten-Leue) |
| [#6](https://github.com/mixeme/imap-scrub/pull/6) | Attachment date subfolder | [jackr0@442939d](https://github.com/jackr0/imap-scrub/commit/442939d) |
| [#7](https://github.com/mixeme/imap-scrub/pull/7) | Date → sender → metadata attachment folders | [mikulcak@5337895a](https://github.com/mikulcak/imap-scrub/commit/5337895a) |
| [#8](https://github.com/mixeme/imap-scrub/pull/8) | Local + macOS build scripts | [mikulcak@efc8fa72](https://github.com/mikulcak/imap-scrub/commit/efc8fa72) |
| [#9](https://github.com/mixeme/imap-scrub/pull/9) | Native `UidMove`; drop `go-imap-move` | [touste@395f2e0](https://github.com/touste/imap-scrub/commit/395f2e0) |
| [#10](https://github.com/mixeme/imap-scrub/pull/10) | Nix flake + direnv | [ahbk@4da9393](https://github.com/ahbk/imap-scrub/commit/4da9393) |
| [#11](https://github.com/mixeme/imap-scrub/pull/11) | `export_mailbox` → local mbox | [aheissenberger@05ace3ee](https://github.com/aheissenberger/imap-scrub/commit/05ace3ee) |
| [#12](https://github.com/mixeme/imap-scrub/pull/12) | VS Code / Cursor Go 1.20 dev container | [mikulcak@e2c84719](https://github.com/mikulcak/imap-scrub/commit/e2c84719) |
| [#13](https://github.com/mixeme/imap-scrub/pull/13) | Continuation branding, `mixeme` install/updater paths | this release |

### Already in upstream 0.0.6 (no separate port PR)

- Skip `remove_attachments` when a message has no attachments ([michalfapso#7](https://github.com/axllent/imap-scrub/pull/7))
- Mailbox `Select()` hang fix (aheissenberger `fix-select-mailbox-blocked`, in `lib/mailboxes.go`)

### Highlights

- Fix `older_than` / add `newer_than` using IMAP **internal delivery date**, with local midnight cutoffs and post-fetch re-checks
- `pass_file`, comma-separated `from`, richer attachment save paths
- `export_mailbox` action for mbox backup
- Native IMAP MOVE, Nix flake, local/macOS/Windows build scripts, CI + optional IMAP integration tests
- Self-updater and docs point at `mixeme/imap-scrub`


## [0.0.6]

- Avoid altering messages that don't have attachments (#7) thanks to @michalfapso


## [0.0.5]

- Write full output paths of saved attachments into "%d-attachments-deleted.txt"
- Bump minimum Go version, update Go modules
- Replace deprecated ioutil references


## [0.0.4]

- Fix confusing include_unread output (#5)
- Code cleanup
- Add GitHub actions workflow
- New internal self-updater
- Update Go modules


## [0.0.3]

- Upgrade modules / dependencies


## [0.0.2]

- Support darwin arm64
- Remove ineffectual assignment to inlineClosed
- Handle trash detection errors


## [0.0.1]

- Initial release
