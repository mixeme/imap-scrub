# Changelog

## [Unreleased]

- Fix `older_than` to use the IMAP internal delivery date instead of the sender-controlled `Date` header (ported from axllent/imap-scrub#14)
- Normalize the cutoff to local midnight and re-verify each fetched message before delete or attachment actions
- Clarify debug output when messages are skipped: show internal date and cutoff in local time, and note when the envelope `Date` header differs
- Add local build scripts for Windows and Linux amd64 binaries (from axllent/imap-scrub#14)
- Add a Docker-based Linux build script
- Add `pass_file` config option to load IMAP password from a file (ported from axllent/imap-scrub#13, thanks @Carsten-Leue)
- Support comma-separated `from` filter with OR matching (ported from axllent/imap-scrub#13, thanks @Carsten-Leue)
- Add `newer_than` day filter using IMAP internal date (ported from axllent/imap-scrub#13, adapted from SentSince; thanks @Carsten-Leue)
- Save attachments under a per-message date subfolder `DD-Mon-YY` (ported from jackr0/imap-scrub@442939d, thanks @jackr0)
- Organize saved attachments as date → sender → per-email metadata folders (recipient, subject, UID) (ported from mikulcak/imap-scrub@5337895a, adapted hierarchy; thanks @mikulcak)
- Add local `scripts/build.sh` and macOS cross-compile `scripts/build-macos.sh` (ported from mikulcak/imap-scrub@efc8fa72 / @08a74437, thanks @mikulcak)
- Add unit tests and GitHub Actions CI for core date, search, and pass_file logic
- Add CI checks that local build scripts produce Windows and Linux amd64 binaries
- Add optional IMAP integration tests and mailbox seeding script (`testdata/imap/`)
- Use native `UidMove` from `go-imap` and drop the `go-imap-move` dependency (ported from touste/imap-scrub@395f2e0, thanks @touste)
- Add Nix flake (`flake.nix`) and direnv `.envrc` for reproducible builds (ported from ahbk/imap-scrub@4da9393, thanks @ahbk)


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
