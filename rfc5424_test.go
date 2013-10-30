package syslogparser

import (
	. "launchpad.net/gocheck"
	"time"
)

type Rfc5424TestSuite struct {
}

var _ = Suite(&Rfc5424TestSuite{})

func (s *Rfc5424TestSuite) TestParseTimestamp_UTC(c *C) {
	buff := []byte("1985-04-12T23:20:50.52Z")
	start := 0
	ts := time.Date(1985, time.April, 12, 23, 20, 50, 52*10e6, time.UTC)

	s.assertTimestamp(c, ts, buff, start, 23, nil)
}

func (s *Rfc5424TestSuite) TestParseTimestamp_NumericTimezone(c *C) {
	tz := "-04:00"
	buff := []byte("1985-04-12T19:20:50.52" + tz)
	start := 0

	tmpTs, err := time.Parse("-07:00", tz)
	c.Assert(err, IsNil)

	ts := time.Date(1985, time.April, 12, 19, 20, 50, 52*10e6, tmpTs.Location())

	s.assertTimestamp(c, ts, buff, start, len(buff), nil)
}

func (s *Rfc5424TestSuite) TestParseTimestamp_MilliSeconds(c *C) {
	buff := []byte("2003-10-11T22:14:15.003Z")
	start := 0

	ts := time.Date(2003, time.October, 11, 22, 14, 15, 3*10e5, time.UTC)

	s.assertTimestamp(c, ts, buff, start, len(buff), nil)
}

func (s *Rfc5424TestSuite) TestParseTimestamp_MicroSeconds(c *C) {
	tz := "-07:00"
	buff := []byte("2003-08-24T05:14:15.000003" + tz)
	start := 0

	tmpTs, err := time.Parse("-07:00", tz)
	c.Assert(err, IsNil)

	ts := time.Date(2003, time.August, 24, 5, 14, 15, 3*10e2, tmpTs.Location())

	s.assertTimestamp(c, ts, buff, start, len(buff), nil)
}

func (s *Rfc5424TestSuite) TestParseTimestamp_NanoSeconds(c *C) {
	buff := []byte("2003-08-24T05:14:15.000000003-07:00")
	start := 0
	ts := new(time.Time)

	s.assertTimestamp(c, *ts, buff, start, 26, ErrTimestampUnknownFormat)
}

func (s *Rfc5424TestSuite) TestFindNextSpace_NoSpace(c *C) {
	buff := []byte("aaaaaa")
	start := 0

	s.assertFindNextSpace(c, 0, buff, start, ErrNoSpace)
}

func (s *Rfc5424TestSuite) TestFindNextSpace_SpaceFound(c *C) {
	buff := []byte("foo bar baz")
	start := 0

	s.assertFindNextSpace(c, 4, buff, start, nil)
}

func (s *Rfc5424TestSuite) TestParseYear_Invalid(c *C) {
	buff := []byte("1a2b")
	start := 0
	expected := 0

	s.assertParseYear(c, expected, buff, start, 4, ErrYearInvalid)
}

func (s *Rfc5424TestSuite) TestParseYear_TooShort(c *C) {
	buff := []byte("123")
	start := 0
	expected := 0

	s.assertParseYear(c, expected, buff, start, 0, ErrEOL)
}

func (s *Rfc5424TestSuite) TestParseYear_Valid(c *C) {
	buff := []byte("2013")
	start := 0
	expected := 2013

	s.assertParseYear(c, expected, buff, start, 4, nil)
}

func (s *Rfc5424TestSuite) TestParseMonth_InvalidString(c *C) {
	buff := []byte("ab")
	start := 0
	expected := 0

	s.assertParseMonth(c, expected, buff, start, 2, ErrMonthInvalid)
}

func (s *Rfc5424TestSuite) TestParseMonth_InvalidRange(c *C) {
	buff := []byte("00")
	start := 0
	expected := 0

	s.assertParseMonth(c, expected, buff, start, 2, ErrMonthInvalid)

	// ----

	buff = []byte("13")

	s.assertParseMonth(c, expected, buff, start, 2, ErrMonthInvalid)
}

func (s *Rfc5424TestSuite) TestParseMonth_TooShort(c *C) {
	buff := []byte("1")
	start := 0
	expected := 0

	s.assertParseMonth(c, expected, buff, start, 0, ErrEOL)
}

func (s *Rfc5424TestSuite) TestParseMonth_Valid(c *C) {
	buff := []byte("02")
	start := 0
	expected := 2

	s.assertParseMonth(c, expected, buff, start, 2, nil)
}

func (s *Rfc5424TestSuite) TestParseDay_InvalidString(c *C) {
	buff := []byte("ab")
	start := 0
	expected := 0

	s.assertParseDay(c, expected, buff, start, 2, ErrDayInvalid)
}

func (s *Rfc5424TestSuite) TestParseDay_TooShort(c *C) {
	buff := []byte("1")
	start := 0
	expected := 0

	s.assertParseDay(c, expected, buff, start, 0, ErrEOL)
}

func (s *Rfc5424TestSuite) TestParseDay_InvalidRange(c *C) {
	buff := []byte("00")
	start := 0
	expected := 0

	s.assertParseDay(c, expected, buff, start, 2, ErrDayInvalid)

	// ----

	buff = []byte("32")

	s.assertParseDay(c, expected, buff, start, 2, ErrDayInvalid)
}

func (s *Rfc5424TestSuite) TestParseDay_Valid(c *C) {
	buff := []byte("02")
	start := 0
	expected := 2

	s.assertParseDay(c, expected, buff, start, 2, nil)
}

func (s *Rfc5424TestSuite) TestParseFullDate_Invalid(c *C) {
	buff := []byte("2013+10-28")
	start := 0
	fd := rfc5424FullDate{}

	s.assertParseFullDate(c, fd, buff, start, 4, ErrTimestampUnknownFormat)

	// ---

	buff = []byte("2013-10+28")
	s.assertParseFullDate(c, fd, buff, start, 7, ErrTimestampUnknownFormat)
}

func (s *Rfc5424TestSuite) TestParseFullDate_Valid(c *C) {
	buff := []byte("2013-10-28")
	start := 0
	fd := rfc5424FullDate{
		year:  2013,
		month: 10,
		day:   28,
	}

	s.assertParseFullDate(c, fd, buff, start, len(buff), nil)
}

func (s *Rfc5424TestSuite) TestParseHour_InvalidString(c *C) {
	buff := []byte("azer")
	start := 0
	expected := 0

	s.assertParseHour(c, expected, buff, start, 2, ErrHourInvalid)
}

func (s *Rfc5424TestSuite) TestParseHour_TooShort(c *C) {
	buff := []byte("1")
	start := 0
	expected := 0

	s.assertParseHour(c, expected, buff, start, 0, ErrEOL)
}

func (s *Rfc5424TestSuite) TestParseHour_InvalidRange(c *C) {
	buff := []byte("-1")
	start := 0
	expected := 0

	s.assertParseHour(c, expected, buff, start, 2, ErrHourInvalid)

	// ----

	buff = []byte("24")

	s.assertParseHour(c, expected, buff, start, 2, ErrHourInvalid)
}

func (s *Rfc5424TestSuite) TestParseHour_Valid(c *C) {
	buff := []byte("12")
	start := 0
	expected := 12

	s.assertParseHour(c, expected, buff, start, 2, nil)
}

func (s *Rfc5424TestSuite) TestParseMinute_InvalidString(c *C) {
	buff := []byte("azer")
	start := 0
	expected := 0

	s.assertParseMinute(c, expected, buff, start, 2, ErrMinuteInvalid)
}

func (s *Rfc5424TestSuite) TestParseMinute_TooShort(c *C) {
	buff := []byte("1")
	start := 0
	expected := 0

	s.assertParseMinute(c, expected, buff, start, 0, ErrEOL)
}

func (s *Rfc5424TestSuite) TestParseMinute_InvalidRange(c *C) {
	buff := []byte("-1")
	start := 0
	expected := 0

	s.assertParseMinute(c, expected, buff, start, 2, ErrMinuteInvalid)

	// ----

	buff = []byte("60")

	s.assertParseMinute(c, expected, buff, start, 2, ErrMinuteInvalid)
}

func (s *Rfc5424TestSuite) TestParseMinute_Valid(c *C) {
	buff := []byte("12")
	start := 0
	expected := 12

	s.assertParseMinute(c, expected, buff, start, 2, nil)
}

func (s *Rfc5424TestSuite) TestParseSecond_InvalidString(c *C) {
	buff := []byte("azer")
	start := 0
	expected := 0

	s.assertParseSecond(c, expected, buff, start, 2, ErrSecondInvalid)
}

func (s *Rfc5424TestSuite) TestParseSecond_TooShort(c *C) {
	buff := []byte("1")
	start := 0
	expected := 0

	s.assertParseSecond(c, expected, buff, start, 0, ErrEOL)
}

func (s *Rfc5424TestSuite) TestParseSecond_InvalidRange(c *C) {
	buff := []byte("-1")
	start := 0
	expected := 0

	s.assertParseSecond(c, expected, buff, start, 2, ErrSecondInvalid)

	// ----

	buff = []byte("60")

	s.assertParseSecond(c, expected, buff, start, 2, ErrSecondInvalid)
}

func (s *Rfc5424TestSuite) TestParseSecond_Valid(c *C) {
	buff := []byte("12")
	start := 0
	expected := 12

	s.assertParseSecond(c, expected, buff, start, 2, nil)
}

func (s *Rfc5424TestSuite) TestParseSecFrac_InvalidString(c *C) {
	buff := []byte("azerty")
	start := 0
	expected := 0.0

	s.assertParseSecFrac(c, expected, buff, start, 0, ErrSecFracInvalid)
}

func (s *Rfc5424TestSuite) TestParseSecFrac_NanoSeconds(c *C) {
	buff := []byte("123456789")
	start := 0
	expected := 0.123456

	s.assertParseSecFrac(c, expected, buff, start, 6, nil)
}

func (s *Rfc5424TestSuite) TestParseSecFrac_Valid(c *C) {
	buff := []byte("0")
	start := 0

	expected := 0.0
	s.assertParseSecFrac(c, expected, buff, start, 1, nil)

	buff = []byte("52")
	expected = 0.52
	s.assertParseSecFrac(c, expected, buff, start, 2, nil)

	buff = []byte("003")
	expected = 0.003
	s.assertParseSecFrac(c, expected, buff, start, 3, nil)

	buff = []byte("000003")
	expected = 0.000003
	s.assertParseSecFrac(c, expected, buff, start, 6, nil)
}

func (s *Rfc5424TestSuite) TestParseNumericalTimeOffset_Valid(c *C) {
	buff := []byte("+02:00")
	cursor := 0
	l := len(buff)
	tmpTs, err := time.Parse("-07:00", string(buff))
	c.Assert(err, IsNil)

	obtained, err := parseNumericalTimeOffset(buff, &cursor, l)
	c.Assert(err, IsNil)

	expected := tmpTs.Location()
	c.Assert(obtained, DeepEquals, expected)

	c.Assert(cursor, Equals, 6)
}

func (s *Rfc5424TestSuite) TestParseTimeOffset_Valid(c *C) {
	buff := []byte("Z")
	cursor := 0
	l := len(buff)

	obtained, err := parseTimeOffset(buff, &cursor, l)
	c.Assert(err, IsNil)
	c.Assert(obtained, DeepEquals, time.UTC)
	c.Assert(cursor, Equals, 1)
}

func (s *Rfc5424TestSuite) TestGetHourMin_Valid(c *C) {
	buff := []byte("12:34")
	cursor := 0
	l := len(buff)

	expectedHour := 12
	expectedMinute := 34

	obtainedHour, obtainedMinute, err := getHourMinute(buff, &cursor, l)
	c.Assert(err, IsNil)
	c.Assert(obtainedHour, Equals, expectedHour)
	c.Assert(obtainedMinute, Equals, expectedMinute)

	c.Assert(cursor, Equals, l)
}

func (s *Rfc5424TestSuite) TestParsePartialTime_Valid(c *C) {
	buff := []byte("05:14:15.000003")
	cursor := 0
	l := len(buff)

	obtained, err := parsePartialTime(buff, &cursor, l)
	expected := rfc5424PartialTime{
		hour:    5,
		minute:  14,
		seconds: 15,
		secFrac: 0.000003,
	}

	c.Assert(err, IsNil)
	c.Assert(obtained, DeepEquals, expected)
	c.Assert(cursor, Equals, l)
}

func (s *Rfc5424TestSuite) TestParseFullTime_Valid(c *C) {
	tz := "-02:00"
	buff := []byte("05:14:15.000003" + tz)
	cursor := 0
	l := len(buff)

	tmpTs, err := time.Parse("-07:00", string(tz))
	c.Assert(err, IsNil)

	obtainedFt, err := parseFullTime(buff, &cursor, l)
	expectedFt := rfc5424FullTime{
		pt: rfc5424PartialTime{
			hour:    5,
			minute:  14,
			seconds: 15,
			secFrac: 0.000003,
		},
		loc: tmpTs.Location(),
	}

	c.Assert(err, IsNil)
	c.Assert(obtainedFt, DeepEquals, expectedFt)
	c.Assert(cursor, Equals, 21)
}

func (s *Rfc5424TestSuite) TestToNSec(c *C) {
	fixtures := []float64{
		0.52,
		0.003,
		0.000003,
	}

	expected := []int{
		520000000,
		3000000,
		3000,
	}

	c.Assert(len(fixtures), Equals, len(expected))
	for i, f := range fixtures {
		obtained, err := toNSec(f)
		c.Assert(err, IsNil)
		c.Assert(obtained, Equals, expected[i])
	}
}

// -------------

func (s *Rfc5424TestSuite) assertTimestamp(c *C, ts time.Time, b []byte, cursor int, expC int, e error) {
	p := newRfc5424Parser(b, cursor, len(b))
	obtained, err := p.parseTimestamp()
	c.Assert(err, Equals, e)

	tFmt := time.RFC3339Nano
	c.Assert(obtained.Format(tFmt), Equals, ts.Format(tFmt))

	c.Assert(p.cursor, Equals, expC)
}

func (s *Rfc5424TestSuite) assertFindNextSpace(c *C, nextSpace int, b []byte, from int, e error) {
	obtained, err := findNextSpace(b, from, len(b))
	c.Assert(obtained, Equals, nextSpace)
	c.Assert(err, Equals, e)
}

func (s *Rfc5424TestSuite) assertParseYear(c *C, year int, b []byte, cursor int, expC int, e error) {
	obtained, err := parseYear(b, &cursor, len(b))
	c.Assert(obtained, Equals, year)
	c.Assert(err, Equals, e)
	c.Assert(cursor, Equals, expC)
}

func (s *Rfc5424TestSuite) assertParseMonth(c *C, month int, b []byte, cursor int, expC int, e error) {
	obtained, err := parseMonth(b, &cursor, len(b))
	c.Assert(obtained, Equals, month)
	c.Assert(err, Equals, e)
	c.Assert(cursor, Equals, expC)
}

func (s *Rfc5424TestSuite) assertParseDay(c *C, day int, b []byte, cursor int, expC int, e error) {
	obtained, err := parseDay(b, &cursor, len(b))
	c.Assert(obtained, Equals, day)
	c.Assert(err, Equals, e)
	c.Assert(cursor, Equals, expC)
}

func (s *Rfc5424TestSuite) assertParseFullDate(c *C, fd rfc5424FullDate, b []byte, cursor int, expC int, e error) {
	obtained, err := parseFullDate(b, &cursor, len(b))
	c.Assert(err, Equals, e)
	c.Assert(obtained, Equals, fd)
	c.Assert(cursor, Equals, expC)
}

func (s *Rfc5424TestSuite) assertParseHour(c *C, hour int, b []byte, cursor int, expC int, e error) {
	obtained, err := parseHour(b, &cursor, len(b))
	c.Assert(obtained, Equals, hour)
	c.Assert(err, Equals, e)
	c.Assert(cursor, Equals, expC)
}

func (s *Rfc5424TestSuite) assertParseMinute(c *C, minute int, b []byte, cursor int, expC int, e error) {
	obtained, err := parseMinute(b, &cursor, len(b))
	c.Assert(obtained, Equals, minute)
	c.Assert(err, Equals, e)
	c.Assert(cursor, Equals, expC)
}

func (s *Rfc5424TestSuite) assertParseSecond(c *C, second int, b []byte, cursor int, expC int, e error) {
	obtained, err := parseSecond(b, &cursor, len(b))
	c.Assert(obtained, Equals, second)
	c.Assert(err, Equals, e)
	c.Assert(cursor, Equals, expC)
}

func (s *Rfc5424TestSuite) assertParseSecFrac(c *C, secFrac float64, b []byte, cursor int, expC int, e error) {
	obtained, err := parseSecFrac(b, &cursor, len(b))
	c.Assert(obtained, Equals, secFrac)
	c.Assert(err, Equals, e)
	c.Assert(cursor, Equals, expC)
}
