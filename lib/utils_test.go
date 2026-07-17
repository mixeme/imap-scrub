package lib

import (
	"os"
	"path"
	"testing"
	"time"
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
