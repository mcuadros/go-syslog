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
	buff := []byte("Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8")
	start := 0
	l := len(buff)

	p := newRfc3164Parser(buff, start, l)
	expectedP := &rfc3164Parser{
		buff:   buff,
		cursor: start,
		l:      l,
	}

	c.Assert(p, DeepEquals, expectedP)

	err := p.parse()
	c.Assert(err, IsNil)

	now := time.Now()

	obtained := p.dump()
	expected := logParts{
		"timestamp": time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC),
		"hostname":  "mymachine",
		"tag":       "su",
		"content":   "'su root' failed for lonvick on /dev/pts/8",
	}

	c.Assert(obtained, DeepEquals, expected)
}

func (s *Rfc3164TestSuite) TestParseHeader_Valid(c *C) {
	buff := []byte("Oct 11 22:14:15 mymachine ")
	start := 0
	now := time.Now()
	hdr := rfc3164Header{
		timestamp: time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC),
		hostname:  "mymachine",
	}

	s.assertRfc3164Header(c, hdr, buff, start, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseHeader_InvalidTimestamp(c *C) {
	buff := []byte("Oct 34 32:72:82 mymachine ")
	start := 0
	hdr := rfc3164Header{}

	s.assertRfc3164Header(c, hdr, buff, start, 16, ErrTimestampUnknownFormat)
}

func (s *Rfc3164TestSuite) TestParseMessage_Valid(c *C) {
	content := "foo bar baz blah quux"
	buff := []byte("sometag[123]: " + content)
	start := 0
	hdr := rfc3164Message{
		tag:     "sometag",
		content: content,
	}

	s.assertRfc3164Message(c, hdr, buff, start, len(buff), ErrEOL)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_TooLong(c *C) {
	// XXX : <15 chars
	buff := []byte("aaa")
	start := 0
	ts := new(time.Time)

	s.assertTimestamp(c, *ts, buff, start, len(buff), ErrEOL)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_Invalid(c *C) {
	buff := []byte("Oct 34 32:72:82")
	start := 0
	ts := new(time.Time)

	s.assertTimestamp(c, *ts, buff, start, len(buff), ErrTimestampUnknownFormat)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_TrailingSpace(c *C) {
	// XXX : no year specified. Assumed current year
	// XXX : no timezone specified. Assume UTC
	buff := []byte("Oct 11 22:14:15 ")
	start := 0

	now := time.Now()
	ts := time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC)

	s.assertTimestamp(c, ts, buff, start, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_Valid(c *C) {
	// XXX : no year specified. Assumed current year
	// XXX : no timezone specified. Assume UTC
	buff := []byte("Oct 11 22:14:15")
	start := 0

	now := time.Now()
	ts := time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC)

	s.assertTimestamp(c, ts, buff, start, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseHostname_Invalid(c *C) {
	// XXX : no year specified. Assumed current year
	// XXX : no timezone specified. Assume UTC
	buff := []byte("host name")
	start := 0
	hostname := "host"

	s.assertHostname(c, hostname, buff, start, len("host")+1, nil)
}

func (s *Rfc3164TestSuite) TestParseHostname_Valid(c *C) {
	// XXX : no year specified. Assumed current year
	// XXX : no timezone specified. Assume UTC
	buff := []byte("ubuntu11.somehost.com")
	start := 0
	hostname := string(buff)

	s.assertHostname(c, hostname, buff, start, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTag_TooLong(c *C) {
	// The TAG is a string of ABNF alphanumeric characters that MUST NOT exceed 32 characters.
	// Source : http://tools.ietf.org/html/rfc3164#section-4.1.3

	aaa := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	buff := []byte(aaa + "[10]:")
	start := 0
	tag := ""

	s.assertTag(c, tag, buff, start, len(aaa)+1, ErrTagTooLong)
}

func (s *Rfc3164TestSuite) TestParseTag_Pid(c *C) {
	buff := []byte("apache2[10]:")
	start := 0
	tag := "apache2"

	s.assertTag(c, tag, buff, start, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTag_NoPid(c *C) {
	buff := []byte("apache2:")
	start := 0
	tag := "apache2"

	s.assertTag(c, tag, buff, start, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTag_TrailingSpace(c *C) {
	buff := []byte("apache2: ")
	start := 0
	tag := "apache2"

	s.assertTag(c, tag, buff, start, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseContent_Valid(c *C) {
	buff := []byte(" foo bar baz quux ")
	start := 0
	content := string(bytes.Trim(buff, " "))

	p := newRfc3164Parser(buff, start, len(buff))
	obtained, err := p.parseContent()
	c.Assert(err, Equals, ErrEOL)
	c.Assert(obtained, Equals, content)
	c.Assert(p.cursor, Equals, len(content))
}

func (s *Rfc3164TestSuite) BenchmarkParseTimestamp(c *C) {
	buff := []byte("Oct 11 22:14:15")
	l := len(buff)

	p := newRfc3164Parser(buff, 0, l)

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
	var start int
	l := len(buff)

	p := newRfc3164Parser(buff, start, l)

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
	l := len(buff)

	p := newRfc3164Parser(buff, 0, l)

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
	l := len(buff)

	p := newRfc3164Parser(buff, 0, l)

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
	var start int
	l := len(buff)

	p := newRfc3164Parser(buff, start, l)

	for i := 0; i < c.N; i++ {
		_, err := p.parseMessage()
		if err != ErrEOL {
			panic(err)
		}

		p.cursor = 0
	}
}

func (s *Rfc3164TestSuite) assertTimestamp(c *C, ts time.Time, b []byte, cursor int, expC int, e error) {
	p := newRfc3164Parser(b, cursor, len(b))
	obtained, err := p.parseTimestamp()
	c.Assert(obtained, Equals, ts)
	c.Assert(p.cursor, Equals, expC)
	c.Assert(err, Equals, e)
}

func (s *Rfc3164TestSuite) assertHostname(c *C, h string, b []byte, cursor int, expC int, e error) {
	p := newRfc3164Parser(b, cursor, len(b))
	obtained, err := p.parseHostname()
	c.Assert(obtained, Equals, h)
	c.Assert(p.cursor, Equals, expC)
	c.Assert(err, Equals, e)
}

func (s *Rfc3164TestSuite) assertTag(c *C, t string, b []byte, cursor int, expC int, e error) {
	p := newRfc3164Parser(b, cursor, len(b))
	obtained, err := p.parseTag()
	c.Assert(obtained, Equals, t)
	c.Assert(p.cursor, Equals, expC)
	c.Assert(err, Equals, e)
}

func (s *Rfc3164TestSuite) assertRfc3164Header(c *C, hdr rfc3164Header, b []byte, cursor int, expC int, e error) {
	p := newRfc3164Parser(b, cursor, len(b))
	obtained, err := p.parseHeader()
	c.Assert(err, Equals, e)
	c.Assert(obtained, Equals, hdr)
	c.Assert(p.cursor, Equals, expC)
}

func (s *Rfc3164TestSuite) assertRfc3164Message(c *C, msg rfc3164Message, b []byte, cursor int, expC int, e error) {
	p := newRfc3164Parser(b, cursor, len(b))
	obtained, err := p.parseMessage()
	c.Assert(err, Equals, e)
	c.Assert(obtained, Equals, msg)
	c.Assert(p.cursor, Equals, expC)
}
