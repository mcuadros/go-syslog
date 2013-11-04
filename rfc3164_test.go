package syslogparser

import (
	"bytes"
	. "launchpad.net/gocheck"
	"time"
)

type Rfc3164TestSuite struct {
}

var _ = Suite(&Rfc3164TestSuite{})

func (s *Rfc3164TestSuite) TestRfc3164Parser_Valid(c *C) {
	buff := []byte("<34>Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8")

	p := NewRfc3164Parser(buff)
	expectedP := &rfc3164Parser{
		buff:   buff,
		cursor: 0,
		l:      len(buff),
	}

	c.Assert(p, DeepEquals, expectedP)

	err := p.Parse()
	c.Assert(err, IsNil)

	now := time.Now()

	obtained := p.Dump()
	expected := LogParts{
		"timestamp": time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC),
		"hostname":  "mymachine",
		"tag":       "su",
		"content":   "'su root' failed for lonvick on /dev/pts/8",
		"priority":  34,
		"facility":  4,
		"severity":  2,
	}

	c.Assert(obtained, DeepEquals, expected)
}

func (s *Rfc3164TestSuite) TestParseHeader_Valid(c *C) {
	buff := []byte("Oct 11 22:14:15 mymachine ")
	now := time.Now()
	hdr := rfc3164Header{
		timestamp: time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC),
		hostname:  "mymachine",
	}

	s.assertRfc3164Header(c, hdr, buff, 25, nil)
}

func (s *Rfc3164TestSuite) TestParseHeader_InvalidTimestamp(c *C) {
	buff := []byte("Oct 34 32:72:82 mymachine ")
	hdr := rfc3164Header{}

	s.assertRfc3164Header(c, hdr, buff, 16, ErrTimestampUnknownFormat)
}

func (s *Rfc3164TestSuite) TestParseMessage_Valid(c *C) {
	content := "foo bar baz blah quux"
	buff := []byte("sometag[123]: " + content)
	hdr := rfc3164Message{
		tag:     "sometag",
		content: content,
	}

	s.assertRfc3164Message(c, hdr, buff, len(buff), ErrEOL)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_TooLong(c *C) {
	// XXX : <15 chars
	buff := []byte("aaa")
	ts := new(time.Time)

	s.assertTimestamp(c, *ts, buff, len(buff), ErrEOL)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_Invalid(c *C) {
	buff := []byte("Oct 34 32:72:82")
	ts := new(time.Time)

	s.assertTimestamp(c, *ts, buff, len(buff), ErrTimestampUnknownFormat)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_TrailingSpace(c *C) {
	// XXX : no year specified. Assumed current year
	// XXX : no timezone specified. Assume UTC
	buff := []byte("Oct 11 22:14:15 ")

	now := time.Now()
	ts := time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC)

	s.assertTimestamp(c, ts, buff, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_Valid(c *C) {
	// XXX : no year specified. Assumed current year
	// XXX : no timezone specified. Assume UTC
	buff := []byte("Oct 11 22:14:15")

	now := time.Now()
	ts := time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC)

	s.assertTimestamp(c, ts, buff, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTag_TooLong(c *C) {
	// The TAG is a string of ABNF alphanumeric characters that MUST NOT exceed 32 characters.
	// Source : http://tools.ietf.org/html/rfc3164#section-4.1.3

	aaa := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	buff := []byte(aaa + "[10]:")
	tag := ""

	s.assertTag(c, tag, buff, len(aaa)+1, ErrTagTooLong)
}

func (s *Rfc3164TestSuite) TestParseTag_Pid(c *C) {
	buff := []byte("apache2[10]:")
	tag := "apache2"

	s.assertTag(c, tag, buff, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTag_NoPid(c *C) {
	buff := []byte("apache2:")
	tag := "apache2"

	s.assertTag(c, tag, buff, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTag_TrailingSpace(c *C) {
	buff := []byte("apache2: ")
	tag := "apache2"

	s.assertTag(c, tag, buff, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseContent_Valid(c *C) {
	buff := []byte(" foo bar baz quux ")
	content := string(bytes.Trim(buff, " "))

	p := NewRfc3164Parser(buff)
	obtained, err := p.parseContent()
	c.Assert(err, Equals, ErrEOL)
	c.Assert(obtained, Equals, content)
	c.Assert(p.cursor, Equals, len(content))
}

func (s *Rfc3164TestSuite) BenchmarkParseTimestamp(c *C) {
	buff := []byte("Oct 11 22:14:15")

	p := NewRfc3164Parser(buff)

	for i := 0; i < c.N; i++ {
		_, err := p.parseTimestamp()
		if err != nil {
			panic(err)
		}

		p.cursor = 0
	}
}

func (s *Rfc3164TestSuite) BenchmarkParseHostname(c *C) {
	buff := []byte("gimli.local")

	p := NewRfc3164Parser(buff)

	for i := 0; i < c.N; i++ {
		_, err := p.parseHostname()
		if err != nil {
			panic(err)
		}

		p.cursor = 0
	}
}

func (s *Rfc3164TestSuite) BenchmarkParseTag(c *C) {
	buff := []byte("apache2[10]:")

	p := NewRfc3164Parser(buff)

	for i := 0; i < c.N; i++ {
		_, err := p.parseTag()
		if err != nil {
			panic(err)
		}

		p.cursor = 0
	}
}

func (s *Rfc3164TestSuite) BenchmarkParseHeader(c *C) {
	buff := []byte("Oct 11 22:14:15 mymachine ")

	p := NewRfc3164Parser(buff)

	for i := 0; i < c.N; i++ {
		_, err := p.parseHeader()
		if err != nil {
			panic(err)
		}

		p.cursor = 0
	}
}

func (s *Rfc3164TestSuite) BenchmarkParseMessage(c *C) {
	buff := []byte("sometag[123]: foo bar baz blah quux")

	p := NewRfc3164Parser(buff)

	for i := 0; i < c.N; i++ {
		_, err := p.parseMessage()
		if err != ErrEOL {
			panic(err)
		}

		p.cursor = 0
	}
}

func (s *Rfc3164TestSuite) assertTimestamp(c *C, ts time.Time, b []byte, expC int, e error) {
	p := NewRfc3164Parser(b)
	obtained, err := p.parseTimestamp()
	c.Assert(obtained, Equals, ts)
	c.Assert(p.cursor, Equals, expC)
	c.Assert(err, Equals, e)
}

func (s *Rfc3164TestSuite) assertTag(c *C, t string, b []byte, expC int, e error) {
	p := NewRfc3164Parser(b)
	obtained, err := p.parseTag()
	c.Assert(obtained, Equals, t)
	c.Assert(p.cursor, Equals, expC)
	c.Assert(err, Equals, e)
}

func (s *Rfc3164TestSuite) assertRfc3164Header(c *C, hdr rfc3164Header, b []byte, expC int, e error) {
	p := NewRfc3164Parser(b)
	obtained, err := p.parseHeader()
	c.Assert(err, Equals, e)
	c.Assert(obtained, Equals, hdr)
	c.Assert(p.cursor, Equals, expC)
}

func (s *Rfc3164TestSuite) assertRfc3164Message(c *C, msg rfc3164Message, b []byte, expC int, e error) {
	p := NewRfc3164Parser(b)
	obtained, err := p.parseMessage()
	c.Assert(err, Equals, e)
	c.Assert(obtained, Equals, msg)
	c.Assert(p.cursor, Equals, expC)
}
