package lib

import (
	"bytes"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap"
)

func TestSaveAttachmentDateSenderMetadata(t *testing.T) {
	dir := t.TempDir()
	orig := Config.SavePath
	Config.SavePath = dir
	defer func() { Config.SavePath = orig }()

	ts := time.Date(2024, 5, 30, 14, 45, 0, 0, time.UTC)
	origin := AttachmentOrigin{
		Timestamp: ts,
		Recipient: "me@example.com",
		Subject:   "Invoice #42 / Q2",
		UID:       99,
		Mailbox:   "INBOX",
	}
	out, err := SaveAttachment([]byte("hello"), "User@Example.COM", "doc.pdf", origin)
	if err != nil {
		t.Fatalf("SaveAttachment() error = %v", err)
	}

	// date → sender → metadata → file
	metaDir := path.Base(path.Dir(out))
	senderDir := path.Base(path.Dir(path.Dir(out)))
	dateDir := path.Base(path.Dir(path.Dir(path.Dir(out))))

	if dateDir != "2024-05-30" {
		t.Fatalf("path = %q, want date dir 2024-05-30, got %q", out, dateDir)
	}
	if senderDir != "user@example.com" {
		t.Fatalf("path = %q, want sender dir user@example.com, got %q", out, senderDir)
	}
	wantMeta := "to-me@example.com__subj-invoice-42-q2__uid-99"
	if metaDir != wantMeta {
		t.Fatalf("path = %q, want metadata dir %q, got %q", out, wantMeta, metaDir)
	}
	if path.Base(out) != "2cf24d-doc.pdf" {
		// sha256("hello")[:3] hex = 2cf24d
		t.Fatalf("path = %q, want hashed 2cf24d-doc.pdf", out)
	}
	if !FileExists(out) {
		t.Fatalf("expected file to exist: %s", out)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("file contents = %q, want hello", data)
	}
}

func TestSanitizePathSegment(t *testing.T) {
	got := sanitizePathSegment("  Hello World!! ", "fallback", 48)
	if got != "hello-world" {
		t.Fatalf("sanitizePathSegment = %q, want hello-world", got)
	}
	got = sanitizePathSegment("!!!", "fallback", 48)
	if got != "fallback" {
		t.Fatalf("sanitizePathSegment empty = %q, want fallback", got)
	}
}

func TestExportMailboxAction(t *testing.T) {
	r := Rule{Actions: "export_mailbox"}
	if !r.ExportMailbox() {
		t.Fatal("ExportMailbox() = false, want true")
	}
	if r.Delete() || r.SaveAttachments() || r.RemoveAttachments() {
		t.Fatal("export_mailbox should not imply other actions")
	}

	r = Rule{Actions: "save_attachments, export_mailbox"}
	if !r.ExportMailbox() || !r.SaveAttachments() {
		t.Fatal("combined actions should enable both flags")
	}
}

func TestCreateMBOXAndExportMessage(t *testing.T) {
	dir := t.TempDir()
	orig := Config.SavePath
	Config.SavePath = dir
	defer func() { Config.SavePath = orig }()

	mboxFile, err := CreateMBOX("Archive/Client")
	if err != nil {
		t.Fatalf("CreateMBOX() error = %v", err)
	}

	wantPath := path.Join(dir, "Archive", "Client", "mbox")
	if !FileExists(wantPath) {
		t.Fatalf("expected mbox file at %s", wantPath)
	}

	if _, err := CreateMBOX("Archive/Client"); err == nil {
		t.Fatal("CreateMBOX() expected error when file exists")
	}

	raw := "From: alice@example.com\r\nTo: bob@example.com\r\nSubject: Hello\r\n\r\nBody text\r\n"
	msg := &imap.Message{
		Envelope: &imap.Envelope{
			From: []*imap.Address{{MailboxName: "alice", HostName: "example.com"}},
		},
		InternalDate: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Body: map[*imap.BodySectionName]imap.Literal{
			&imap.BodySectionName{}: bytes.NewReader([]byte(raw)),
		},
	}

	if err := ExportMessage(msg, mboxFile.Writer); err != nil {
		t.Fatalf("ExportMessage() error = %v", err)
	}
	if err := mboxFile.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	data, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "From alice@example.com") {
		t.Fatalf("mbox missing From line, got:\n%s", content)
	}
	if !strings.Contains(content, "Subject: Hello") {
		t.Fatalf("mbox missing message headers, got:\n%s", content)
	}
	if !strings.Contains(content, "Body text") {
		t.Fatalf("mbox missing body, got:\n%s", content)
	}
}

func TestExportMessageNilSafe(t *testing.T) {
	if err := ExportMessage(nil, nil); err == nil {
		t.Fatal("ExportMessage(nil) expected error")
	}

	dir := t.TempDir()
	orig := Config.SavePath
	Config.SavePath = dir
	defer func() { Config.SavePath = orig }()

	mboxFile, err := CreateMBOX("INBOX")
	if err != nil {
		t.Fatalf("CreateMBOX() error = %v", err)
	}
	defer mboxFile.Close()

	if err := ExportMessage(&imap.Message{}, mboxFile.Writer); err == nil {
		t.Fatal("ExportMessage with empty body expected error")
	}
}
