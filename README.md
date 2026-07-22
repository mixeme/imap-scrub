# IMAP-Scrub

![IMAP-Scrub Logo](docs/assets/imap-scrub-logo-3-erased.png)

> **Actively maintained continuation of [axllent/imap-scrub](https://github.com/axllent/imap-scrub).**  
> Upstream last released `0.0.6` (Apr 2024). This fork consolidates community ports and continues development: [`mixeme/imap-scrub`](https://github.com/mixeme/imap-scrub).

A command-line utility (Linux, Mac & Windows) to reduce the size of your IMAP mailbox through a series of pre-defined rules. Each rule contains a series of search modifiers, and one or two actions (`delete`, `remove_attachments`, `save_attachments`, `export_mailbox`).

Typical use cases include:

- Freeing mailbox space by removing large attachments from old messages while keeping the email text and headers
- Saving attachments to disk before stripping them from the server (or exporting matching messages to a local mbox)
- Deleting mail that is no longer useful after a set period — for example social notifications, newsletters, or automated alerts
- Targeting specific senders, subjects, or size thresholds so only the rules you define are applied

Changelog: [docs/CHANGELOG.md](docs/CHANGELOG.md)


## Installing

Download the [latest binary release](https://github.com/mixeme/imap-scrub/releases/latest) for your system.

You can also update an existing install with:

```
imap-scrub -u
```

### Homebrew (Linux & macOS)

The formula lives in this repo, so tap it directly (no separate `homebrew-*` repo needed):

```sh
brew tap mixeme/imap-scrub https://github.com/mixeme/imap-scrub
brew install imap-scrub
```

To build the latest unreleased `develop` branch instead of the last tagged release, pass `--HEAD`:

```sh
brew install --HEAD imap-scrub
```


## Usage options

```
Usage: imap-scrub [options] <config.yml>

Options:
  -y, --yes            do actions (based on config rule actions)
  -m, --mailboxes      list mailboxes on server (helpful for configuration)
  -p, --print-config   print config
  -u, --update         update to latest release version
  -v, --version        show app version
```

Without `-y` / `--yes`, IMAP-Scrub only lists matching messages (dry run). Pass `-y` to apply the configured actions.


## Configuration

Each mail account should have a yaml configuration file. IMAP-Scrub does not currently support OAUTH, so username/password IMAP login is required.

For Gmail with 2-Step Verification, use an [App Password](https://support.google.com/accounts/answer/185833) instead of your normal account password.

### Example config

```yaml
name: My Gmail Account
host: imap.gmail.com
user: example-user@gmail.com
pass: MySecretPassword123
# Alternatively, use pass_file to load password from a file:
# pass_file: /path/to/password-file.txt
save_path: /home/me/email-files
rules:
  - mailbox: "[Gmail]/All Mail"  # IMAP mailbox name
    min_size: 5120               # minimum size in kB
    older_than: 365              # days
    actions: remove_attachments
  - mailbox: "[Gmail]/All Mail"
    from: invitations@linkedin.com, updates@linkedin.com
    older_than: 30
    actions: delete
  - mailbox: "[Gmail]/All Mail"
    from: myclient@example.com
    min_size: 512
    older_than: 90
    newer_than: 30
    actions: save_attachments, remove_attachments
  - mailbox: "[Gmail]/All Mail"
    from: archive-me@example.com
    older_than: 365
    actions: export_mailbox
```

See [All yaml config options](#all-yaml-config-options) below for more info.


## All yaml config options

```yaml
name:        string # reference name of this account
host:        string # IMAP hostname
ssl:         true   # use SSL (default true)
port:        993    # IMAP port number (default 993 if SSL is true, else 143)
user:        string # IMAP username
pass:        string # IMAP password (use either pass or pass_file)
pass_file:   string # path to file containing IMAP password (optional, takes precedence over pass)
save_path:   string # local directory to save attachments and mbox exports (default current dir)
export_path: string # local directory for export_mailbox mbox files (default: save_path)
use_trash:   false  # see below
rules:
  - mailbox:         string # IMAP mailbox name, or '*'/'%' wildcard pattern matching several (see below)
    min_size:        0      # minimum message size in kB
    older_than:      0      # older than x days (see below)
    newer_than:      0      # newer than x days (see below)
    from:            string # match "From" field, comma-separated for OR matching
    to:              string # match "To" field
    subject:         string # match email subject
    body:            string # match email body
    text:            string # match email message
    actions:         string # see below
    include_unread:  false  # include unread messages (default false)
    include_starred: false  # include starred messages (default false)
    keep_signatures: true   # preserve S/MIME signed messages on remove_attachments (default true)
```


### Option: `pass_file`

The `pass_file` option allows you to store your IMAP password in a separate file instead of directly in the configuration file. This is useful for security purposes, as it allows you to keep sensitive credentials separate from your configuration.

When `pass_file` is specified, the password will be read from the file (with whitespace trimmed) and will take precedence over the `pass` field. If the file cannot be read, the application will exit with an error.

**Security recommendation:** Set the directory containing the password file to mode `0700` and the password file itself to mode `0600` to ensure only the owner can access it:

```bash
chmod 0700 /home/me/.secrets
chmod 0600 /home/me/.secrets/gmail-password.txt
```

Example:

```yaml
name: My Gmail Account
host: imap.gmail.com
user: example-user@gmail.com
pass_file: /home/me/.secrets/gmail-password.txt
```


### Option: `from`

The `from` option matches the email `From` header.

You can provide either a single sender or a comma-separated list. When multiple senders are provided, a message matches if it matches any one of them.

Example:

```yaml
from: billing@example.com, invoices@example.com, receipts@example.com
```


### Option: `mailbox`

The mailbox you wish to search. On standard IMAP servers this is probably `INBOX`.

On Gmail this is possibly `[Gmail]/All Mail` or `[Google Mail]/All Mail`, but may differ based on your selected language. To list the mailboxes on your IMAP server to make a choice, run `imap-scrub -m <your-config.yml>` which will print out all mailboxes in your account.

`mailbox` also accepts IMAP's own wildcards so one rule can apply to many folders: `*` matches zero or more characters including the hierarchy delimiter, and `%` matches zero or more characters but stops at the delimiter (e.g. `INBOX.*` matches `INBOX` and every folder below it; `INBOX.%` matches only its direct children). Non-selectable mailboxes (e.g. Gmail's `[Gmail]` parent) are skipped automatically.

Example:

```yaml
- mailbox: "INBOX.*"  # every folder under INBOX, any depth
  older_than: 90
  actions: delete
```


### Option: `use_trash`

If `use_trash` is set to `true`, and your IMAP returns a trash mailbox, then deleted messages will be moved into this mailbox. **Note** that Gmail does not support IMAP delete, so `use_trash` will always be set to `true` for Gmail.


### Option: `include_unread` / `include_starred`

By default both are `false`, so rules only match **read** and **unstarred** messages:

- `include_unread: false` — skip unread (`\Seen` required)
- `include_starred: false` — skip starred / flagged (`\Flagged` excluded)

Set either to `true` when you want that class of messages included in the search.


### Option: `actions`

There are four possible actions, namely:

- `save_attachments` will save any attachments under a date → sender → per-email metadata layout (date from the message `Date` header):

  `save_path/<YYYY-MM-DD>/<sender>/<to-<recipient>__subj-<short-subject>__uid-<uid>>/<hash>-<filename>`
- `remove_attachments` will remove the all attachments and inline images from the original email
- `delete` will simply delete the email
- `export_mailbox` will write matching messages to a local `mbox` file under `export_path/<mailbox-path>/mbox` (`save_path` is used if `export_path` isn't set; nested IMAP mailbox names become directories)

The `actions:` config may include a combination of `save_attachments` and one other (comma-separated), eg :`actions: save_attachments, remove_attachments`.

`export_mailbox` can be used alone or combined with other actions (for example export then `delete`).

**Note** that you cannot combine `remove_attachments` and `delete`.


### Option: `export_path`

By default `export_mailbox` writes its `mbox` files under `save_path`, alongside saved attachments. Set `export_path` to send exports to a separate directory instead:

```yaml
save_path: /home/me/attachments
export_path: /home/me/mailbox-backups
```

If the mbox file for a mailbox already exists, `export_mailbox` **appends** to it instead of failing, and skips any message whose `Message-Id` is already in the file. This makes it safe to rerun the same rule repeatedly (e.g. from a cron job) without duplicating messages already exported.

Without `-y` (dry run), IMAP-Scrub still checks the mbox file (read-only) and reports how many matching messages would be newly exported vs. already present, so you can preview a resumed export before applying it.


### Option: `keep_signatures`

By default (`keep_signatures: true`), `remove_attachments` skips S/MIME signed messages entirely rather than stripping their signature parts.

An S/MIME signed message is either an opaque `application/pkcs7-mime` message (`smime.p7m` / `smime.p7z`, [RFC 8551](https://www.rfc-editor.org/rfc/rfc8551)) or a `multipart/signed` message with a `smime.p7s` signature part. Rewriting either to remove attachments would invalidate the signature, so imap-scrub leaves the message untouched on the server and logs that it was skipped.

Set `keep_signatures: false` to disable this and fall back to the previous behaviour, where `smime.p7s` / `smime.p7m` / `smime.p7z` parts are treated like any other attachment and removed.


### Option: `older_than` / `newer_than`

`older_than` and `newer_than` are optional day-based filters measured from local midnight today. Both use the IMAP **internal delivery date** (when the server stored the message), not the sender's `Date` header. This is intentional: the `Date` header can be wrong, missing, or out of sync with when mail actually arrived.

- `older_than: 30` matches messages delivered before local midnight 30 days ago
- `newer_than: 7` matches messages delivered on or after local midnight 7 days ago
- using both together matches a date range, for example `older_than: 90` with `newer_than: 30` matches messages between 30 and 90 days old

The IMAP `BEFORE` / `SINCE` searches are date-only (no time or timezone), so the server may return messages near the cutoff. IMAP-Scrub re-checks each result locally before listing or changing anything.

Skipped messages near the cutoff are logged with the internal date and cutoff in your local timezone. If the envelope `Date` header falls on a different calendar day, a second line explains the difference.

Example: with `older_than: 3` on 18 June, the cutoff is 15 June 00:00 local. A message whose internal date is 15 June 01:21 local is kept (not old enough), even if its `Date` header says 11 June.


## Development

For building from source, Nix, the dev container, and related notes, see [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).
