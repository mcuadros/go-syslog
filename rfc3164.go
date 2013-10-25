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

func parseHostname(buff []byte, cursor *int, l int) (string, error) {
	from := *cursor
	var to int

	for to = from; to < l; to++ {
		if buff[to] == ' ' {
			break
		}
	}

	hostname := buff[from:to]

	*cursor += to

	// XXX : Start for the next parser
	if *cursor < l {
		*cursor++
	}

	return string(hostname), nil
}

// http://tools.ietf.org/html/rfc3164#section-4.1.3
func parseTag(buff []byte, cursor *int, l int) (string, error) {
	var b byte
	var endOfTag bool
	var bracketOpen bool
	var tag []byte
	var err error
	var found bool
	var tooLong bool

	from := *cursor
	maxLen := from + 32

	for {
		b = buff[*cursor]
		bracketOpen = (b == '[')
		endOfTag = (b == ':' || b == ' ')
		tooLong = (*cursor > maxLen)

		if tooLong {
			return "", ErrTagTooLong
		}

		// XXX : parse PID ?
		if bracketOpen {
			tag = buff[from:*cursor]
			found = true
		}

		if endOfTag {
			if !found {
				tag = buff[from:*cursor]
				found = true
			}

			*cursor++
			break
		}

		*cursor++
	}

	return string(tag), err
}

func parseContent(buff []byte, cursor *int, l int) (string, error) {
	if *cursor > l {
		return "", ErrEOL
	}

	return string(buff[*cursor:l]), nil
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
