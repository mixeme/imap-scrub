# IMAP integration test fixtures

Synthetic `.eml` messages for manual and integration testing of imap-scrub.

## Fixtures

| File | Purpose |
|------|---------|
| `01-old-sender-a-with-attachment.eml` | `older_than`, `from`, attachments |
| `02-recent-sender-b.eml` | comma-separated `from` (OR) |
| `03-recent-sender-a-no-attachment.eml` | recent sender-a without attachments |
| `04-unread-sender-a.eml` | `include_unread` default behaviour |
| `05-starred-sender-b.eml` | `include_starred` default behaviour |
| `06-recent-internal-old-envelope.eml` | `older_than` / `newer_than` use internal date, not Date header |

Subjects are prefixed with `[imap-scrub-test]` so they are easy to find in the mailbox.

## Seed the test mailbox

1. Copy `testdata/imap/imap-test.yml.example` to `dev/imap-test.yml` and fill in credentials.
2. Store the password in `dev/imap-test.pass` and reference it via `pass_file`.
3. Run:

```sh
python scripts/seed-imap-test.py
```

The script regenerates fixtures under `testdata/imap/fixtures/` and APPENDs them to `INBOX`
with appropriate IMAP flags and internal dates.

Credentials live only under `dev/` (gitignored).

## Automated tests

Unit tests run on every PR via GitHub Actions (`go test ./...`).

Integration tests (real IMAP mailbox) are optional:

```sh
go test -tags=integration ./lib -run TestIntegration -count=1
```

Requires `dev/imap-test.yml` (see `imap-test.yml.example`). Re-seed fixtures with
`python scripts/seed-imap-test.py` if messages are missing.

`TestIntegrationExportMailbox` fetches a fixture with `BODY.PEEK[]`, writes it via
`export_mailbox` helpers, and checks the resulting `mbox` file.
