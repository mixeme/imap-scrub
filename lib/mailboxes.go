package lib

import (
	"sort"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// ListMailboxes returns a list of Mailboxes on the server
func ListMailboxes(cReader *client.Client) {
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- cReader.List("", "*", mailboxes)
	}()

	Log.InfoF("Mailboxes on %s\n", Config.Name)
	for m := range mailboxes {
		if !InStringSlice("\\Noselect", m.Attributes) {
			Log.Info(" - " + m.Name)
		}
	}

	if err := <-done; err != nil {
		Log.ErrorF("%v\n", err)
	}
}

// IsMailboxPattern reports whether a rule's mailbox uses IMAP LIST wildcards
// ('*' matches zero or more characters including hierarchy delimiters, '%'
// matches zero or more characters except the hierarchy delimiter — RFC 3501
// section 6.3.8) rather than naming a single mailbox.
func IsMailboxPattern(mailbox string) bool {
	return strings.ContainsAny(mailbox, "*%")
}

// ExpandMailboxPattern resolves a rule's mailbox to the concrete, selectable
// mailbox names it applies to. A plain name (no '*' / '%') is returned as-is,
// so a typo still surfaces as the usual "no such mailbox" error from Select
// instead of silently matching nothing. A pattern is expanded via IMAP LIST,
// using the server's own wildcard semantics, skipping \Noselect mailboxes
// (e.g. Gmail's "[Gmail]" parent), and returned sorted for deterministic
// processing order.
func ExpandMailboxPattern(cReader *client.Client, mailbox string) ([]string, error) {
	if !IsMailboxPattern(mailbox) {
		return []string{mailbox}, nil
	}

	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- cReader.List("", mailbox, mailboxes)
	}()

	names := []string{}
	for m := range mailboxes {
		if InStringSlice("\\Noselect", m.Attributes) {
			continue
		}
		names = append(names, m.Name)
	}

	if err := <-done; err != nil {
		return nil, err
	}

	sort.Strings(names)
	return names, nil
}

// DetectTrash will return the trash folder of a Gmail account, if applicable
// Gmail only supports moving to the trash
func DetectTrash(cReader *client.Client) (string, error) {
	if !Config.UseTrash && Config.Host != "imap.gmail.com" {
		return "", nil
	}

	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- cReader.List("", "*", mailboxes)
	}()

	var trashMailbox = ""
	for m := range mailboxes {
		if InStringSlice("\\Trash", m.Attributes) {
			Log.DebugF("Deleted messages will be moved to \"%s\"", m.Name)
			trashMailbox = m.Name
		}
	}

	if err := <-done; err != nil {
		Log.ErrorF("%v\n", err)
	}

	if trashMailbox != "" {
		return trashMailbox, nil
	}

	return "", nil
}
