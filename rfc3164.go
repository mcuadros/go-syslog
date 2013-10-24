package syslogparser

import (
	"time"
)

// https://tools.ietf.org/html/rfc3164#section-4.1.2
func parseTimestamp(buff []byte, cursor *int, l int) (time.Time, error) {
	var ts time.Time

	tsFmt := "Jan 02 15:04:05"
	// len(fmt)
	tsFmtLen := 15

	if *cursor+tsFmtLen > l {
		*cursor = l
		return ts, ErrEOL
	}

	sub := buff[*cursor:tsFmtLen]
	ts, err := time.Parse(tsFmt, string(sub))
	if err != nil {
		// XXX : where to move the cursor in this situation ?
		*cursor = len(sub)
		return ts, ErrTimestampUnknownFormat
	}

	fixTimestampIfNeeded(&ts)

	*cursor += 15
	return ts, nil
}

func fixTimestampIfNeeded(ts *time.Time) {
	now := time.Now()
	y := ts.Year()

	if ts.Year() == 0 {
		y = now.Year()
	}

	newTs := time.Date(y, ts.Month(), ts.Day(), ts.Hour(), ts.Minute(),
		ts.Second(), ts.Nanosecond(), ts.Location())

	*ts = newTs
}
