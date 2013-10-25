package syslogparser

import (
	"bytes"
	"time"
)

type rfc3164Parser struct {
	buff    []byte
	cursor  *int
	l       int
	header  rfc3164Header
	message rfc3164Message
}

func newRfc3164Parser(buff []byte, cursor *int, l int) *rfc3164Parser {
	return &rfc3164Parser{
		buff:   buff,
		cursor: cursor,
		l:      l,
	}
}

func (p *rfc3164Parser) parse() error {
	hdr, err := parseHeader(p.buff, p.cursor, p.l)
	if err != nil {
		return err
	}

	msg, err := parseMessage(p.buff, p.cursor, p.l)
	if err != ErrEOL {
		return err
	}

	p.header = hdr
	p.message = msg

	return nil
}

func (p *rfc3164Parser) dump() logParts {
	return logParts{
		"timestamp": p.header.timestamp,
		"hostname":  p.header.hostname,
		"tag":       p.message.tag,
		"content":   p.message.content,
	}
}

func parseHeader(buff []byte, cursor *int, l int) (rfc3164Header, error) {
	hdr := rfc3164Header{}
	var err error

	ts, err := parseTimestamp(buff, cursor, l)
	if err != nil {
		return hdr, err
	}

	hostname, err := parseHostname(buff, cursor, l)
	if err != nil {
		return hdr, err
	}

	hdr.timestamp = ts
	hdr.hostname = hostname

	return hdr, nil
}

func parseMessage(buff []byte, cursor *int, l int) (rfc3164Message, error) {
	msg := rfc3164Message{}
	var err error

	tag, err := parseTag(buff, cursor, l)
	if err != nil {
		return msg, err
	}

	content, err := parseContent(buff, cursor, l)
	if err != ErrEOL {
		return msg, err
	}

	msg.tag = tag
	msg.content = content

	return msg, err
}

// https://tools.ietf.org/html/rfc3164#section-4.1.2
func parseTimestamp(buff []byte, cursor *int, l int) (time.Time, error) {
	var ts time.Time
	var err error

	tsFmt := "Jan 02 15:04:05"
	// len(fmt)
	tsFmtLen := 15

	if *cursor+tsFmtLen > l {
		*cursor = l
		return ts, ErrEOL
	}

	sub := buff[*cursor:tsFmtLen]
	ts, err = time.Parse(tsFmt, string(sub))
	if err != nil {
		*cursor = len(sub)

		// XXX : If the timestamp is invalid we try to push the cursor one byte
		// XXX : further, in case it is a space
		if (*cursor < l) && (buff[*cursor] == ' ') {
			*cursor++
		}

		return ts, ErrTimestampUnknownFormat
	}

	fixTimestampIfNeeded(&ts)

	*cursor += 15

	if (*cursor < l) && (buff[*cursor] == ' ') {
		*cursor++
	}

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

	*cursor = to

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

	if (*cursor < l) && (buff[*cursor] == ' ') {
		*cursor++
	}

	return string(tag), err
}

func parseContent(buff []byte, cursor *int, l int) (string, error) {
	if *cursor > l {
		return "", ErrEOL
	}

	content := bytes.Trim(buff[*cursor:l], " ")
	*cursor += len(content)

	return string(content), ErrEOL
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
