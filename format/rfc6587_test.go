package format

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	. "gopkg.in/check.v1"
)

func (s *FormatSuite) TestRFC6587_GetSplitFuncSingleSplit(c *C) {
	f := RFC6587{}

	buf := strings.NewReader("10 I am test.")
	scanner := bufio.NewScanner(buf)
	scanner.Split(f.GetSplitFunc())

	r := scanner.Scan()
	c.Assert(r, NotNil)
	c.Assert(scanner.Text(), Equals, "I am test.")
}

func (s *FormatSuite) TestRFC6587_GetSplitFuncMultiSplit(c *C) {
	f := RFC6587{}

	find := []string{
		"I am test.",
		"I am test 2.",
		"hahahahah",
	}
	buf := new(bytes.Buffer)
	for _, i := range find {
		fmt.Fprintf(buf, "%d %s", len(i), i)
	}
	scanner := bufio.NewScanner(buf)
	scanner.Split(f.GetSplitFunc())

	i := 0
	for scanner.Scan() {
		c.Assert(scanner.Text(), Equals, find[i])
		i++
	}

	c.Assert(i, Equals, len(find))
}

func (s *FormatSuite) TestRFC6587_GetSplitFuncMultiSplitNonTransparentFraming(c *C) {
	f := RFC6587{}

	find := []string{
		"<1> I am a test.",
		"<2> I am a test 2.",
		"<3> hahahah",
	}
	buf := new(bytes.Buffer)
	for _, i := range find {
		fmt.Fprintf(buf, "%s", i)
	}
	scanner := bufio.NewScanner(buf)
	scanner.Split(f.GetSplitFunc())

	i := 0
	for scanner.Scan() {
		c.Assert(scanner.Text(), Equals, strings.Join(find, ""))
		i++
	}

	c.Assert(i, Equals, 1)
}

func (s *FormatSuite) TestRFC6587_GetSplitBadSplit(c *C) {
	f := RFC6587{}

	find := "I am test.2 ab"
	buf := strings.NewReader("9 " + find)
	scanner := bufio.NewScanner(buf)
	scanner.Split(f.GetSplitFunc())

	r := scanner.Scan()
	c.Assert(r, NotNil)
	c.Assert(scanner.Text(), Equals, find[0:9])

	r = scanner.Scan()
	c.Assert(r, NotNil)

	err := scanner.Err()
	c.Assert(err, ErrorMatches, "strconv.*: parsing \".2\": invalid syntax")
}
