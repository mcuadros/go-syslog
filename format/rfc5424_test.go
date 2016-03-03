package format

import (
	. "gopkg.in/check.v1"
)

func (s *FormatSuite) TestRFC5424_SingleSplit(c *C) {
	f := RFC5424{}
	c.Assert(f.GetSplitFunc(), IsNil)
}
