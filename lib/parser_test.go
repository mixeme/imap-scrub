package lib

import (
	"bytes"
	"strings"
	"testing"

	"github.com/emersion/go-imap"
)

func messageWithBody(raw string) *imap.Message {
	return &imap.Message{
		Envelope: &imap.Envelope{
			From: []*imap.Address{{MailboxName: "alice", HostName: "example.com"}},
			To:   []*imap.Address{{MailboxName: "bob", HostName: "example.com"}},
		},
		Body: map[*imap.BodySectionName]imap.Literal{
			{}: bytes.NewReader([]byte(raw)),
		},
	}
}

const plainAttachmentMessage = "From: alice@example.com\r\n" +
	"To: bob@example.com\r\n" +
	"Subject: Hello\r\n" +
	"MIME-Version: 1.0\r\n" +
	"Content-Type: multipart/mixed; boundary=\"AAA\"\r\n" +
	"\r\n" +
	"--AAA\r\n" +
	"Content-Type: text/plain\r\n" +
	"\r\n" +
	"Hello world\r\n" +
	"--AAA\r\n" +
	"Content-Type: application/octet-stream; name=\"file.bin\"\r\n" +
	"Content-Disposition: attachment; filename=\"file.bin\"\r\n" +
	"Content-Transfer-Encoding: base64\r\n" +
	"\r\n" +
	"aGVsbG8=\r\n" +
	"--AAA--\r\n"

const multipartSignedMessage = "From: alice@example.com\r\n" +
	"To: bob@example.com\r\n" +
	"Subject: Signed\r\n" +
	"MIME-Version: 1.0\r\n" +
	"Content-Type: multipart/signed; protocol=\"application/pkcs7-signature\"; micalg=sha-256; boundary=\"BBB\"\r\n" +
	"\r\n" +
	"--BBB\r\n" +
	"Content-Type: text/plain\r\n" +
	"\r\n" +
	"Signed content\r\n" +
	"--BBB\r\n" +
	"Content-Type: application/pkcs7-signature; name=\"smime.p7s\"\r\n" +
	"Content-Transfer-Encoding: base64\r\n" +
	"Content-Disposition: attachment; filename=\"smime.p7s\"\r\n" +
	"\r\n" +
	"ZmFrZXNpZ25hdHVyZQ==\r\n" +
	"--BBB--\r\n"

const opaquePkcs7MimeMessage = "From: alice@example.com\r\n" +
	"To: bob@example.com\r\n" +
	"Subject: Opaque\r\n" +
	"MIME-Version: 1.0\r\n" +
	"Content-Type: application/pkcs7-mime; smime-type=signed-data; name=\"smime.p7m\"\r\n" +
	"Content-Transfer-Encoding: base64\r\n" +
	"Content-Disposition: attachment; filename=\"smime.p7m\"\r\n" +
	"\r\n" +
	"ZmFrZXA3bQ==\r\n"

func TestHandleMessageRemovesPlainAttachment(t *testing.T) {
	keep := true
	rule := Rule{Mailbox: "INBOX", Actions: "remove_attachments", PreserveSMIME: &keep}

	raw, count, err := HandleMessage(messageWithBody(plainAttachmentMessage), rule)
	if err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("HandleMessage() count = %d, want 1", count)
	}
	if strings.Contains(raw, "file.bin") {
		t.Fatalf("expected attachment to be stripped, got:\n%s", raw)
	}
	if !strings.Contains(raw, "attachments-deleted.txt") {
		t.Fatalf("expected attachments-deleted notice, got:\n%s", raw)
	}
}

func TestHandleMessageSkipsMultipartSignedByDefault(t *testing.T) {
	keep := true
	rule := Rule{Mailbox: "INBOX", Actions: "remove_attachments", PreserveSMIME: &keep}

	_, count, err := HandleMessage(messageWithBody(multipartSignedMessage), rule)
	if err == nil {
		t.Fatal("HandleMessage() expected error (skip) for multipart/signed message, got nil")
	}
	if count != 0 {
		t.Fatalf("HandleMessage() count = %d, want 0", count)
	}
}

func TestHandleMessageSkipsOpaquePkcs7MimeByDefault(t *testing.T) {
	keep := true
	rule := Rule{Mailbox: "INBOX", Actions: "remove_attachments", PreserveSMIME: &keep}

	_, count, err := HandleMessage(messageWithBody(opaquePkcs7MimeMessage), rule)
	if err == nil {
		t.Fatal("HandleMessage() expected error (skip) for opaque application/pkcs7-mime message, got nil")
	}
	if count != 0 {
		t.Fatalf("HandleMessage() count = %d, want 0", count)
	}
}

func TestHandleMessageStripsSignatureWhenKeepSignaturesDisabled(t *testing.T) {
	noKeep := false
	rule := Rule{Mailbox: "INBOX", Actions: "remove_attachments", PreserveSMIME: &noKeep}

	raw, count, err := HandleMessage(messageWithBody(multipartSignedMessage), rule)
	if err != nil {
		t.Fatalf("HandleMessage() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("HandleMessage() count = %d, want 1", count)
	}
	if strings.Contains(raw, "smime.p7s") {
		t.Fatalf("expected smime.p7s to be stripped when keep_signatures is false, got:\n%s", raw)
	}
}
