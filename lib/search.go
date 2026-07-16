package lib

import (
	"fmt"
	"net/textproto"
	"strings"

	"github.com/emersion/go-imap"
)

// ParseFromSenders splits a comma-separated from filter into trimmed sender terms.
func ParseFromSenders(from string) []string {
	terms := []string{}
	for _, sender := range strings.Split(from, ",") {
		sender = strings.TrimSpace(sender)
		if sender == "" {
			continue
		}
		terms = append(terms, sender)
	}
	return terms
}

// UnionUIDs returns the unique union of UID slices preserving arbitrary order.
func UnionUIDs(idGroups ...[]uint32) []uint32 {
	seen := map[uint32]struct{}{}
	for _, group := range idGroups {
		for _, id := range group {
			seen[id] = struct{}{}
		}
	}
	result := make([]uint32, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	return result
}

// FromSearchCriteria builds one or more IMAP search criteria for comma-separated from.
// When from is empty, a single copy of base is returned.
func FromSearchCriteria(base imap.SearchCriteria, from string) ([]*imap.SearchCriteria, string) {
	terms := ParseFromSenders(from)
	if len(terms) == 0 {
		criteria := base
		return []*imap.SearchCriteria{&criteria}, ""
	}

	criteriaList := make([]*imap.SearchCriteria, 0, len(terms))
	for _, sender := range terms {
		senderCriteria := base
		senderCriteria.Header = make(textproto.MIMEHeader, len(base.Header)+1)
		for key, values := range base.Header {
			senderCriteria.Header[key] = append([]string{}, values...)
		}
		senderCriteria.Header["From"] = []string{sender}
		criteriaCopy := senderCriteria
		criteriaList = append(criteriaList, &criteriaCopy)
	}

	return criteriaList, fmt.Sprintf("from any of: \"%s\"", strings.Join(terms, "\", \""))
}
