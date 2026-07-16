package lib

import (
	"os"
	"path"
	"testing"
	"time"
)

func TestSaveAttachmentDateSubfolder(t *testing.T) {
	dir := t.TempDir()
	orig := Config.SavePath
	Config.SavePath = dir
	defer func() { Config.SavePath = orig }()

	ts := time.Date(2024, 5, 30, 14, 45, 0, 0, time.UTC)
	out, err := SaveAttachment([]byte("hello"), "user@example.com", "doc.pdf", ts)
	if err != nil {
		t.Fatalf("SaveAttachment() error = %v", err)
	}

	wantDate := "30-May-24"
	if path.Base(path.Dir(out)) != wantDate {
		t.Fatalf("path = %q, want parent dir %q", out, wantDate)
	}
	if path.Base(path.Dir(path.Dir(out))) != "user@example.com" {
		t.Fatalf("path = %q, want sender dir user@example.com", out)
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
