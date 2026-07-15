package lib

import (
	"time"

	"github.com/emersion/go-imap"
)

// BeginningOfDay returns local midnight for t, preserving its location.
// older_than cutoffs are anchored to the user's calendar day, not UTC.
func BeginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// MessageIsOlderThan returns true when msg was delivered strictly before cutoff.
// cutoff is always local midnight from BeginningOfDay().
func MessageIsOlderThan(msg *imap.Message, cutoff time.Time) bool {
	if msg.InternalDate.IsZero() {
		Log.WarningF("Skipping UID %d: message has no internal date", msg.Uid)
		return false
	}

	if !msg.InternalDate.Before(cutoff) {
		loc := cutoff.Location()
		const localFmt = "2006-01-02 15:04:05 MST"
		internalLocal := msg.InternalDate.In(loc)
		cutoffLocal := cutoff.In(loc)
		Log.DebugF(
			"Skipping UID %d: internal date %s is not before older_than cutoff %s",
			msg.Uid,
			internalLocal.Format(localFmt),
			cutoffLocal.Format(localFmt),
		)
		if msg.Envelope != nil && !msg.Envelope.Date.IsZero() {
			envelopeLocal := msg.Envelope.Date.In(loc)
			if envelopeLocal.Format("2006-01-02") != internalLocal.Format("2006-01-02") {
				Log.DebugF(
					"  envelope Date header is %s (older_than uses internal delivery date, not Date header)",
					envelopeLocal.Format(localFmt),
				)
			}
		}
		return false
	}

	return true
}
