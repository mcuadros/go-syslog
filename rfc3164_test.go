package syslogparser

import (
	. "launchpad.net/gocheck"
	"time"
)

type Rfc3164TestSuite struct {
}

var _ = Suite(&Rfc3164TestSuite{})

func (s *Rfc3164TestSuite) TestParseTimestamp_TooLong(c *C) {
	// XXX : <15 chars
	buff := []byte("aaa")
	start := 0
	ts := new(time.Time)

	assertTimestamp(c, *ts, buff, start, len(buff), ErrEOL)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_Invalid(c *C) {
	buff := []byte("Oct 34 32:72:82")
	start := 0
	ts := new(time.Time)

	assertTimestamp(c, *ts, buff, start, len(buff), ErrTimestampUnknownFormat)
}

func (s *Rfc3164TestSuite) TestParseTimestamp_Valid(c *C) {
	// XXX : no year specified. Assumed current year
	// XXX : no timezone specified. Assume UTC
	buff := []byte("Oct 11 22:14:15")
	start := 0

	now := time.Now()
	ts := time.Date(now.Year(), time.October, 11, 22, 14, 15, 0, time.UTC)

	assertTimestamp(c, ts, buff, start, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseHostname_Invalid(c *C) {
	// XXX : no year specified. Assumed current year
	// XXX : no timezone specified. Assume UTC
	buff := []byte("host name")
	start := 0
	hostname := "host"

	assertHostname(c, hostname, buff, start, len("host")+1, nil)
}

func (s *Rfc3164TestSuite) TestParseHostname_Valid(c *C) {
	// XXX : no year specified. Assumed current year
	// XXX : no timezone specified. Assume UTC
	buff := []byte("ubuntu11.somehost.com")
	start := 0
	hostname := string(buff)

	assertHostname(c, hostname, buff, start, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTag_TooLong(c *C) {
	// The TAG is a string of ABNF alphanumeric characters that MUST NOT exceed 32 characters.
	// Source : http://tools.ietf.org/html/rfc3164#section-4.1.3

	aaa := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	buff := []byte(aaa + "[10]:")
	start := 0
	tag := ""

	assertTag(c, tag, buff, start, len(aaa)+1, ErrTagTooLong)
}

func (s *Rfc3164TestSuite) TestParseTag_Pid(c *C) {
	buff := []byte("apache2[10]:")
	start := 0
	tag := "apache2"

	assertTag(c, tag, buff, start, len(buff), nil)
}

func (s *Rfc3164TestSuite) TestParseTag_NoPid(c *C) {
	buff := []byte("apache2:")
	start := 0
	tag := "apache2"

	assertTag(c, tag, buff, start, len(buff), nil)
}

func (s *Rfc3164TestSuite) BenchmarkParseTimestamp(c *C) {
	buff := []byte("Oct 11 22:14:15")
	var start int
	l := len(buff)

	for i := 0; i < c.N; i++ {
		start = 0
		_, err := parseTimestamp(buff, &start, l)
		if err != nil {
			panic(err)
		}
	}
}

func (s *Rfc3164TestSuite) BenchmarkParseHostname(c *C) {
	buff := []byte("gimli.local")
	var start int
	l := len(buff)

	for i := 0; i < c.N; i++ {
		start = 0
		_, err := parseHostname(buff, &start, l)
		if err != nil {
			panic(err)
		}
	}
}

func (s *Rfc3164TestSuite) BenchmarkParseTag(c *C) {
	buff := []byte("apache2[10]:")
	var start int
	l := len(buff)

	for i := 0; i < c.N; i++ {
		start = 0
		_, err := parseTag(buff, &start, l)
		if err != nil {
			panic(err)
		}
	}
}

func assertTimestamp(c *C, ts time.Time, b []byte, s int, expS int, e error) {
	obtained, err := parseTimestamp(b, &s, len(b))
	c.Assert(obtained, Equals, ts)
	c.Assert(s, Equals, expS)
	c.Assert(err, Equals, e)
}

func assertHostname(c *C, h string, b []byte, s int, expS int, e error) {
	obtained, err := parseHostname(b, &s, len(b))
	c.Assert(obtained, Equals, h)
	c.Assert(s, Equals, expS)
	c.Assert(err, Equals, e)
}

func assertTag(c *C, t string, b []byte, s int, expS int, e error) {
	obtained, err := parseTag(b, &s, len(b))
	c.Assert(obtained, Equals, t)
	c.Assert(s, Equals, expS)
	c.Assert(err, Equals, e)
}
