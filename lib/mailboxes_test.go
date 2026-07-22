package lib

import "testing"

func TestIsMailboxPattern(t *testing.T) {
	cases := map[string]bool{
		"INBOX":         false,
		"Archive/2024":  false,
		"INBOX.*":       true,
		"INBOX.%":       true,
		"INBOX.Sub*box": true,
	}
	for mailbox, want := range cases {
		if got := IsMailboxPattern(mailbox); got != want {
			t.Errorf("IsMailboxPattern(%q) = %v, want %v", mailbox, got, want)
		}
	}
}
