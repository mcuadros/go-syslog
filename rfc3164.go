package syslogparser

import (
	"bytes"
	"time"
)

func NewRfc3164Parser(buff []byte, cursor int, l int) *rfc3164Parser {
	return &rfc3164Parser{
		buff:   buff,
		cursor: cursor,
		l:      l,
	}
}

func (p *rfc3164Parser) Parse() error {
	hdr, err := p.parseHeader()
	if err != nil {
		return err
	}

	p.cursor++

	msg, err := p.parseMessage()
	if err != ErrEOL {
		return err
	}

	p.header = hdr
	p.message = msg

	return nil
}

func (p *rfc3164Parser) Dump() LogParts {
	return LogParts{
		"timestamp": p.header.timestamp,
		"hostname":  p.header.hostname,
		"tag":       p.message.tag,
		"content":   p.message.content,
	}
}

func (p *rfc3164Parser) parseHeader() (rfc3164Header, error) {
	hdr := rfc3164Header{}
	var err error

	ts, err := p.parseTimestamp()
	if err != nil {
		return hdr, err
	}

	hostname, err := p.parseHostname()
	if err != nil {
		return hdr, err
	}

	hdr.timestamp = ts
	hdr.hostname = hostname

	return hdr, nil
}

func (p *rfc3164Parser) parseMessage() (rfc3164Message, error) {
	msg := rfc3164Message{}
	var err error

	tag, err := p.parseTag()
	if err != nil {
		return msg, err
	}

	content, err := p.parseContent()
	if err != ErrEOL {
		return msg, err
	}

	msg.tag = tag
	msg.content = content

	return msg, err
}

// https://tools.ietf.org/html/rfc3164#section-4.1.2
func (p *rfc3164Parser) parseTimestamp() (time.Time, error) {
	var ts time.Time
	var err error

	tsFmt := "Jan 02 15:04:05"
	// len(fmt)
	tsFmtLen := 15

	if p.cursor+tsFmtLen > p.l {
		p.cursor = p.l
		return ts, ErrEOL
	}

	sub := p.buff[p.cursor:tsFmtLen]
	ts, err = time.Parse(tsFmt, string(sub))
	if err != nil {
		p.cursor = len(sub)

		// XXX : If the timestamp is invalid we try to push the cursor one byte
		// XXX : further, in case it is a space
		if (p.cursor < p.l) && (p.buff[p.cursor] == ' ') {
			p.cursor++
		}

		return ts, ErrTimestampUnknownFormat
	}

	fixTimestampIfNeeded(&ts)

	p.cursor += 15

	if (p.cursor < p.l) && (p.buff[p.cursor] == ' ') {
		p.cursor++
	}

	return ts, nil
}

func (p *rfc3164Parser) parseHostname() (string, error) {
	return parseHostname(p.buff, &p.cursor, p.l)
}

// http://tools.ietf.org/html/rfc3164#section-4.1.3
func (p *rfc3164Parser) parseTag() (string, error) {
	var b byte
	var endOfTag bool
	var bracketOpen bool
	var tag []byte
	var err error
	var found bool
	var tooLong bool

	from := p.cursor
	maxLen := from + 32

	for {
		b = p.buff[p.cursor]
		bracketOpen = (b == '[')
		endOfTag = (b == ':' || b == ' ')
		tooLong = (p.cursor > maxLen)

		if tooLong {
			return "", ErrTagTooLong
		}

		// XXX : parse PID ?
		if bracketOpen {
			tag = p.buff[from:p.cursor]
			found = true
		}

		if endOfTag {
			if !found {
				tag = p.buff[from:p.cursor]
				found = true
			}

			p.cursor++
			break
		}

		p.cursor++
	}

	if (p.cursor < p.l) && (p.buff[p.cursor] == ' ') {
		p.cursor++
	}

	return string(tag), err
}

func (p *rfc3164Parser) parseContent() (string, error) {
	if p.cursor > p.l {
		return "", ErrEOL
	}

	content := bytes.Trim(p.buff[p.cursor:p.l], " ")
	p.cursor += len(content)

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
