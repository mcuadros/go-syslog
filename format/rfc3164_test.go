package format

import (
	. "gopkg.in/check.v1"
)

func (s *FormatSuite) TestRFC3164_SingleSplit(c *C) {
	f := RFC3164{}
	c.Assert(f.GetSplitFunc(), IsNil)
}
