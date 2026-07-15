#!/usr/bin/env python3
"""Generate IMAP test fixtures and upload them to the configured test mailbox."""

from __future__ import annotations

import email
import email.policy
import imaplib
import os
import re
import sys
import time
from datetime import datetime, timedelta, timezone
from email.mime.application import MIMEApplication
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
FIXTURES_DIR = ROOT / "testdata" / "imap" / "fixtures"
CONFIG_PATH = ROOT / "dev" / "imap-test.yml"
MAILBOX = "INBOX"


def load_config(path: Path) -> dict[str, str]:
    text = path.read_text(encoding="utf-8")
    config: dict[str, str] = {}
    pass_file: str | None = None
    for line in text.splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        if ":" not in line:
            continue
        key, value = line.split(":", 1)
        key = key.strip()
        value = value.strip().strip('"').strip("'")
        if value.startswith("#"):
            continue
        config[key] = value
        if key == "pass_file":
            pass_file = value

    if pass_file:
        pass_path = ROOT / pass_file
        config["pass"] = pass_path.read_text(encoding="utf-8").strip()
    elif "pass" not in config:
        raise SystemExit("Config must define pass_file or pass")

    for required in ("host", "user", "pass"):
        if not config.get(required):
            raise SystemExit(f"Missing required config value: {required}")

    config["port"] = config.get("port", "993")
    config["ssl"] = config.get("ssl", "true").lower()
    return config


def format_imap_date(dt: datetime) -> str:
    if dt.tzinfo is None:
        dt = dt.astimezone()
    return imaplib.Time2Internaldate(dt)


def build_message(
    *,
    subject: str,
    sender: str,
    body: str,
    sent_at: datetime,
    attachment_name: str | None = None,
    attachment_bytes: bytes | None = None,
) -> bytes:
    if attachment_name and attachment_bytes is not None:
        msg = MIMEMultipart()
        msg.attach(MIMEText(body, "plain", "utf-8"))
        part = MIMEApplication(attachment_bytes, Name=attachment_name)
        part.add_header("Content-Disposition", "attachment", filename=attachment_name)
        msg.attach(part)
    else:
        msg = MIMEText(body, "plain", "utf-8")

    msg["Subject"] = subject
    msg["From"] = sender
    msg["To"] = "test@mixeml.ru"
    msg["Date"] = email.utils.format_datetime(sent_at)
    return msg.as_bytes(policy=email.policy.SMTP)


def fixture_specs(now: datetime) -> list[dict]:
    tiny_pdf = (
        b"%PDF-1.1\n"
        b"1 0 obj<<>>endobj\n"
        b"trailer<<>>\n"
        b"%%EOF\n"
    )
    old = now - timedelta(days=45)
    recent = now - timedelta(days=2)

    return [
        {
            "filename": "01-old-sender-a-with-attachment.eml",
            "subject": "[imap-scrub-test] old message from sender-a with attachment",
            "sender": "sender-a@test.com",
            "body": "Old message for older_than and attachment tests.",
            "sent_at": old,
            "internal_at": old,
            "flags": "(\\Seen)",
            "attachment_name": "test.pdf",
            "attachment_bytes": tiny_pdf,
        },
        {
            "filename": "02-recent-sender-b.eml",
            "subject": "[imap-scrub-test] recent message from sender-b",
            "sender": "sender-b@test.com",
            "body": "Recent message for comma-separated from OR matching.",
            "sent_at": recent,
            "internal_at": recent,
            "flags": "(\\Seen)",
        },
        {
            "filename": "03-recent-sender-a-no-attachment.eml",
            "subject": "[imap-scrub-test] recent message from sender-a without attachment",
            "sender": "sender-a@test.com",
            "body": "Recent sender-a message without attachments.",
            "sent_at": recent,
            "internal_at": recent,
            "flags": "(\\Seen)",
        },
        {
            "filename": "04-unread-sender-a.eml",
            "subject": "[imap-scrub-test] unread message from sender-a",
            "sender": "sender-a@test.com",
            "body": "Unread message for include_unread tests.",
            "sent_at": recent,
            "internal_at": recent,
            "flags": None,
        },
        {
            "filename": "05-starred-sender-b.eml",
            "subject": "[imap-scrub-test] starred message from sender-b",
            "sender": "sender-b@test.com",
            "body": "Starred message for include_starred tests.",
            "sent_at": recent,
            "internal_at": recent,
            "flags": "(\\Seen \\Flagged)",
        },
    ]


def write_fixtures(specs: list[dict]) -> list[tuple[Path, dict]]:
    FIXTURES_DIR.mkdir(parents=True, exist_ok=True)
    written: list[tuple[Path, dict]] = []
    for spec in specs:
        payload = build_message(
            subject=spec["subject"],
            sender=spec["sender"],
            body=spec["body"],
            sent_at=spec["sent_at"],
            attachment_name=spec.get("attachment_name"),
            attachment_bytes=spec.get("attachment_bytes"),
        )
        path = FIXTURES_DIR / spec["filename"]
        path.write_bytes(payload)
        written.append((path, spec))
        print(f"wrote fixture {path.relative_to(ROOT)}")
    return written


def connect(config: dict[str, str]) -> imaplib.IMAP4:
    host = config["host"]
    port = int(config["port"])
    if config.get("ssl", "true") == "true":
        client = imaplib.IMAP4_SSL(host, port)
    else:
        client = imaplib.IMAP4(host, port)
    client.login(config["user"], config["pass"])
    return client


def upload_fixtures(client: imaplib.IMAP4, fixtures: list[tuple[Path, dict]]) -> None:
    for path, spec in fixtures:
        flags = spec.get("flags")
        internal_at = spec["internal_at"]
        date_arg = format_imap_date(internal_at)
        data = path.read_bytes()
        status, response = client.append(
            MAILBOX,
            flags,
            date_arg,
            data,
        )
        if status != "OK":
            raise RuntimeError(f"APPEND failed for {path.name}: {status} {response}")
        print(f"uploaded {path.name} -> {MAILBOX} flags={flags} internal={internal_at.isoformat()}")


def main() -> int:
    if not CONFIG_PATH.exists():
        print(f"Missing config: {CONFIG_PATH}", file=sys.stderr)
        return 1

    config = load_config(CONFIG_PATH)
    now = datetime.now().astimezone()
    specs = fixture_specs(now)
    fixtures = write_fixtures(specs)

    print(f"connecting to {config['host']} as {config['user']} ...")
    client = connect(config)
    try:
        upload_fixtures(client, fixtures)
    finally:
        client.logout()

    print("done")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
