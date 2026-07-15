package lib

import (
	"testing"
	"time"

	"github.com/emersion/go-imap"
)

func TestBeginningOfDay(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	input := time.Date(2026, 7, 15, 14, 30, 0, 0, loc)
	got := BeginningOfDay(input)
	want := time.Date(2026, 7, 15, 0, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Fatalf("BeginningOfDay() = %v, want %v", got, want)
	}
}

func TestMessageIsOlderThan(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	cutoff := time.Date(2026, 7, 10, 0, 0, 0, 0, loc)

	tests := []struct {
		name         string
		internalDate time.Time
		want         bool
	}{
		{
			name:         "strictly older",
			internalDate: time.Date(2026, 7, 9, 23, 59, 0, 0, loc),
			want:         true,
		},
		{
			name:         "on cutoff day",
			internalDate: time.Date(2026, 7, 10, 1, 0, 0, 0, loc),
			want:         false,
		},
		{
			name:         "missing internal date",
			internalDate: time.Time{},
			want:         false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := &imap.Message{
				Uid:          42,
				InternalDate: tc.internalDate,
			}
			if got := MessageIsOlderThan(msg, cutoff); got != tc.want {
				t.Fatalf("MessageIsOlderThan() = %v, want %v", got, tc.want)
			}
		})
	}
}
