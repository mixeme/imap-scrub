package lib

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	mboxlib "github.com/emersion/go-mbox"
)

const (
	maxRecipientSegmentLength = 64
	maxSubjectSegmentLength   = 48
)

// AttachmentOrigin stores metadata used to identify the source email for a saved attachment.
type AttachmentOrigin struct {
	Timestamp time.Time
	Recipient string
	Subject   string
	UID       uint32
	Mailbox   string
}

var resultCount = 1

// PrettyPrint outputs a JSON-encoded representation of an interface
func PrettyPrint(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	fmt.Println(string(s))
}

// CreateDir will check if a directory exists, and create it if not
func CreateDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// PrintHdrDetails returns a IMAP search result
func PrintHdrDetails(msg *imap.Message) {
	e := msg.Envelope
	from := TruncateFromAddress(e.From)
	hrSize := ByteCountSI(msg.Size)
	starred := " "
	if InStringSlice("\\Flagged", msg.Flags) {
		starred = "*"
	}

	Log.InfoF("#%-4d %s  %s %-62s %s%7s", resultCount, e.Date.Format("02-Jan-06"), from, Truncate(e.Subject, 60), starred, hrSize)
	resultCount++
}

// Truncate will return a truncates string
func Truncate(raw string, length int) string {
	var numRunes = 0
	for index := range raw {
		numRunes++
		if numRunes > length {
			return raw[:index-3] + "..."
		}
	}
	return raw
}

// TruncateFromAddress returns a formatted and truncated From address
func TruncateFromAddress(from []*imap.Address) string {
	if len(from) == 0 {
		return "Unknown sender"
	}
	email := fmt.Sprintf(" <%s>", from[0].Address())

	emailLength := len(email)

	remaining := 45 - emailLength

	name := ""
	if remaining > 5 {
		name = Truncate(from[0].PersonalName, remaining)
	}

	return fmt.Sprintf("%-47s", strings.TrimSpace(name+email))
}

// ByteCountSI returns a human-readable size from bytes
func ByteCountSI(b uint32) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

// InStringSlice returns whether a value is in a string slice
func InStringSlice(val string, slice []string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func sanitizePathSegment(raw, fallback string, maxLen int) string {
	raw = strings.TrimSpace(strings.ToLower(raw))

	var b strings.Builder
	b.Grow(len(raw))

	lastDash := false
	for _, r := range raw {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		isAllowedPunctuation := r == '@' || r == '.' || r == '_'

		if isAlphaNum || isAllowedPunctuation {
			b.WriteRune(r)
			lastDash = false
			continue
		}

		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}

	out := strings.Trim(b.String(), "-.")
	if out == "" {
		out = fallback
	}

	if len(out) > maxLen {
		out = strings.Trim(out[:maxLen], "-.")
		if out == "" {
			out = fallback
		}
	}

	return out
}

func emailFolderName(origin AttachmentOrigin) string {
	recipient := sanitizePathSegment(origin.Recipient, "unknown-recipient", maxRecipientSegmentLength)
	subject := sanitizePathSegment(origin.Subject, "no-subject", maxSubjectSegmentLength)

	uidPart := "unknown"
	if origin.UID > 0 {
		uidPart = fmt.Sprintf("%d", origin.UID)
	}

	return fmt.Sprintf("to-%s__subj-%s__uid-%s", recipient, subject, uidPart)
}

// SaveAttachment will save an attachment to
// <outdir>/<YYYY-MM-DD>/<sender>/<to-recipient__subj-subject__uid-uid>/<hash>-<filename>
// returns the output file path and/or error
func SaveAttachment(b []byte, emailAddress, fileName string, origin AttachmentOrigin) (string, error) {
	fileName = path.Clean(filepath.Base(fileName))

	if fileName == "" {
		return "", fmt.Errorf("Filename empty, not saving")
	}

	h := sha256.New()
	h.Write(b)
	hash := h.Sum(nil)

	hashed := fmt.Sprintf("%x-%s", hash[0:3], fileName)

	datePart := "unknown-date"
	if !origin.Timestamp.IsZero() {
		datePart = origin.Timestamp.Format("2006-01-02")
	}

	senderDir := sanitizePathSegment(emailAddress, "no-email", maxRecipientSegmentLength)
	outDir := path.Clean(path.Join(Config.SavePath, datePart, senderDir, emailFolderName(origin)))
	if err := CreateDir(outDir); err != nil {
		return "", err
	}

	outFile := path.Clean(path.Join(outDir, hashed))
	if FileExists(outFile) {
		Log.WarningF(" - %s already exists", outFile)
		return outFile, nil
	}

	// #nosec
	file, err := os.OpenFile(
		outFile,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0664,
	)
	if err != nil {
		return outFile, err
	}
	defer file.Close()

	// Write bytes to file
	bytesWritten, err := file.Write(b)
	if err != nil {
		return outFile, err
	}
	// write a copy of the attachment
	bytes := uint32(bytesWritten)

	// set timestamp
	if !origin.Timestamp.IsZero() {
		_ = os.Chtimes(outFile, origin.Timestamp, origin.Timestamp)
	}

	Log.NoticeF(" - Saved %s (%s)", outFile, ByteCountSI(bytes))

	return outFile, nil
}

// MBOXFile wraps an mbox writer and its underlying file so both can be closed.
type MBOXFile struct {
	Writer *mboxlib.Writer
	file   *os.File
}

// Close finalizes the mbox stream and closes the file.
func (m *MBOXFile) Close() error {
	var firstErr error
	if m.Writer != nil {
		if err := m.Writer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if m.file != nil {
		if err := m.file.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// CreateMBOX creates save_path/<mailbox-path>/mbox for the given IMAP mailbox name.
// Nested mailbox names (e.g. "Archive/2024") become nested directories.
func CreateMBOX(mailboxName string) (*MBOXFile, error) {
	mailboxParts := strings.Split(mailboxName, "/")
	outDir := path.Clean(path.Join(Config.SavePath, path.Join(mailboxParts...)))
	if err := CreateDir(outDir); err != nil {
		return nil, err
	}

	outFile := path.Clean(path.Join(outDir, "mbox"))
	if FileExists(outFile) {
		Log.WarningF(" - File '%s' already exists", outFile)
		return nil, fmt.Errorf("mbox file already exists: %s", outFile)
	}

	// #nosec
	file, err := os.OpenFile(outFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0664)
	if err != nil {
		return nil, err
	}

	return &MBOXFile{
		Writer: mboxlib.NewWriter(file),
		file:   file,
	}, nil
}

// ExportMessage writes a single IMAP message into an mbox writer.
func ExportMessage(msg *imap.Message, mboxWriter *mboxlib.Writer) error {
	if msg == nil {
		return fmt.Errorf("Server didn't returned message")
	}
	if mboxWriter == nil {
		return fmt.Errorf("mbox writer is nil")
	}

	var body io.Reader
	for _, literal := range msg.Body {
		body = literal
		break
	}
	if body == nil {
		return fmt.Errorf("Server didn't returned message body")
	}

	from := "unknown"
	if msg.Envelope != nil && len(msg.Envelope.From) > 0 {
		from = msg.Envelope.From[0].Address()
	}

	w, err := mboxWriter.CreateMessage(from, msg.InternalDate)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, body); err != nil {
		return err
	}

	Log.NoticeF(" - Exported message to local mbox")
	return nil
}

