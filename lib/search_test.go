package lib

import (
	"testing"

	"github.com/emersion/go-imap"
)

func TestParseFromSenders(t *testing.T) {
	got := ParseFromSenders(" a@test.com , , b@test.com ")
	want := []string{"a@test.com", "b@test.com"}
	if len(got) != len(want) {
		t.Fatalf("ParseFromSenders() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParseFromSenders()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestUnionUIDs(t *testing.T) {
	got := UnionUIDs([]uint32{1, 2, 3}, []uint32{2, 4})
	if len(got) != 4 {
		t.Fatalf("UnionUIDs() len = %d, want 4", len(got))
	}
	seen := map[uint32]bool{}
	for _, id := range got {
		seen[id] = true
	}
	for _, id := range []uint32{1, 2, 3, 4} {
		if !seen[id] {
			t.Fatalf("UnionUIDs() missing uid %d", id)
		}
	}
}

func TestFromSearchCriteria(t *testing.T) {
	base := imap.SearchCriteria{}
	criteria, description := FromSearchCriteria(base, "a@test.com, b@test.com")
	if len(criteria) != 2 {
		t.Fatalf("FromSearchCriteria() count = %d, want 2", len(criteria))
	}
	if description == "" {
		t.Fatal("FromSearchCriteria() description is empty")
	}
	if got := criteria[0].Header.Get("From"); got != "a@test.com" {
		t.Fatalf("first From = %q, want a@test.com", got)
	}
	if got := criteria[1].Header.Get("From"); got != "b@test.com" {
		t.Fatalf("second From = %q, want b@test.com", got)
	}
}
