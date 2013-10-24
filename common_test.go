package syslogparser

import (
	. "launchpad.net/gocheck"
	"testing"
)

// Hooks up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

type CommonTestSuite struct {
}

var _ = Suite(&CommonTestSuite{})

func (s *CommonTestSuite) TestParsePriority_Empty(c *C) {
	pri := newPriority(0)
	buff := []byte("")
	start := 0

	assertPriority(c, pri, buff, start, start, ErrPriorityEmpty)
}

func (s *CommonTestSuite) TestParsePriority_NoStart(c *C) {
	pri := newPriority(0)
	buff := []byte("7>")
	start := 0

	assertPriority(c, pri, buff, start, start, ErrPriorityNoStart)
}

func (s *CommonTestSuite) TestParsePriority_NoEnd(c *C) {
	pri := newPriority(0)
	buff := []byte("<77")
	start := 0

	assertPriority(c, pri, buff, start, start, ErrPriorityNoEnd)
}

func (s *CommonTestSuite) TestParsePriority_TooShort(c *C) {
	pri := newPriority(0)
	buff := []byte("<>")
	start := 0

	assertPriority(c, pri, buff, start, start, ErrPriorityTooShort)
}

func (s *CommonTestSuite) TestParsePriority_TooLong(c *C) {
	pri := newPriority(0)
	buff := []byte("<1233>")
	start := 0

	assertPriority(c, pri, buff, start, start, ErrPriorityTooLong)
}

func (s *CommonTestSuite) TestParsePriority_NoDigits(c *C) {
	pri := newPriority(0)
	buff := []byte("<7a8>")
	start := 0

	assertPriority(c, pri, buff, start, start, ErrPriorityNonDigit)
}

func (s *CommonTestSuite) TestParsePriority_Ok(c *C) {
	pri := newPriority(190)
	buff := []byte("<190>")
	start := 0

	assertPriority(c, pri, buff, start, start+5, nil)
}

func (s *CommonTestSuite) TestNewPriority(c *C) {
	obtained := newPriority(165)

	expected := priority{
		f: facility{value: 20},
		s: severity{value: 5},
	}

	c.Assert(obtained, DeepEquals, expected)
}

func (s *CommonTestSuite) TestParseVersion_NotFound(c *C) {
	buff := []byte("<123>")
	start := 5

	assertVersion(c, NO_VERSION, buff, start, start, ErrVersionNotFound)
}

func (s *CommonTestSuite) TestParseVersion_NonDigit(c *C) {
	buff := []byte("<123>a")
	start := 5

	assertVersion(c, NO_VERSION, buff, start, start+1, nil)
}

func (s *CommonTestSuite) TestParseVersion_Ok(c *C) {
	buff := []byte("<123>1")
	start := 5

	assertVersion(c, 1, buff, start, start+1, nil)
}

func assertPriority(c *C, p priority, b []byte, s int, expS int, e error) {
	obtained, err := parsePriority(b, &s, len(b))
	c.Assert(obtained, DeepEquals, p)
	c.Assert(s, Equals, expS)
	c.Assert(err, Equals, e)
}

func assertVersion(c *C, version int, b []byte, s int, expS int, e error) {
	obtained, err := parseVersion(b, &s, len(b))
	c.Assert(obtained, Equals, version)
	c.Assert(s, Equals, expS)
	c.Assert(err, Equals, e)
}
