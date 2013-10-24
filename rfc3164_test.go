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

func assertTimestamp(c *C, ts time.Time, b []byte, s int, expS int, e error) {
	obtained, err := parseTimestamp(b, &s, len(b))
	c.Assert(obtained, Equals, ts)
	c.Assert(s, Equals, expS)
	c.Assert(err, Equals, e)
}
