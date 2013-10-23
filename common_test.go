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

	runAsserts(c, pri, buff, start, ErrPriorityEmpty)
}

func (s *CommonTestSuite) TestParsePriority_NoStart(c *C) {
	pri := newPriority(0)
	buff := []byte("7>")
	start := 0

	runAsserts(c, pri, buff, start, ErrPriorityNoStart)
}

func (s *CommonTestSuite) TestParsePriority_NoEnd(c *C) {
	pri := newPriority(0)
	buff := []byte("<77")
	start := 0

	runAsserts(c, pri, buff, start, ErrPriorityNoEnd)
}

func (s *CommonTestSuite) TestParsePriority_TooShort(c *C) {
	pri := newPriority(0)
	buff := []byte("<>")
	start := 0

	runAsserts(c, pri, buff, start, ErrPriorityTooShort)
}

func (s *CommonTestSuite) TestParsePriority_TooLong(c *C) {
	pri := newPriority(0)
	buff := []byte("<1233>")
	start := 0

	runAsserts(c, pri, buff, start, ErrPriorityTooLong)
}

func (s *CommonTestSuite) TestParsePriority_NoDigits(c *C) {
	pri := newPriority(0)
	buff := []byte("<7a8>")
	start := 0

	runAsserts(c, pri, buff, start, ErrPriorityNonDigit)
}

func (s *CommonTestSuite) TestParsePriority_Ok(c *C) {
	pri := newPriority(190)
	buff := []byte("<190>")
	start := 0

	runAsserts(c, pri, buff, start, nil)
}

func (s *CommonTestSuite) TestNewPriority(c *C) {
	obtained := newPriority(165)

	expected := Priority{
		Facility: Facility{Value: 20},
		Severity: Severity{Value: 5},
	}

	c.Assert(obtained, DeepEquals, expected)
}

func runAsserts(c *C, p Priority, b []byte, i int, e error) {
	obtained, err := ParsePriority(b, &i, len(b))
	c.Assert(obtained, DeepEquals, p)
	c.Assert(err, Equals, e)
}
