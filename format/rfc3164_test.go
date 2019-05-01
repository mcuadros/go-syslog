package format

import (
	. "gopkg.in/check.v1"
)

func (s *FormatSuite) TestRFC3164_SingleSplit(c *C) {
	f := RFC3164{}
	c.Assert(f.GetSplitFunc(), IsNil)
}

func (s *FormatSuite) TestRFC3164_CorrectParsingTypical(c *C) {
	f := RFC3164{}

	find := []string{
		`<13>May  1 20:51:40 myhostname myprogram: ciao`,
	}
	parser := f.GetParser([]byte(find[0]))
	err := parser.Parse()
	c.Assert(err, IsNil)
	c.Assert(parser.Dump()["content"], Equals, "ciao")
	c.Assert(parser.Dump()["hostname"], Equals, "myhostname")
	c.Assert(parser.Dump()["tag"], Equals, "myprogram")

}

func (s *FormatSuite) TestRFC3164_CorrectParsingGnu(c *C) {
	// GNU implementation of syslog() has a variant: hostname is missing
	f := RFC3164{}

	find := []string{
		`<13>May  1 20:51:40 myprogram: ciao`,
	}
	parser := f.GetParser([]byte(find[0]))
	err := parser.Parse()
	c.Assert(err, IsNil)
	c.Assert(parser.Dump()["content"], Equals, "ciao")
	// c.Assert(parser.Dump()["hostname"], Equals, "myhostname")
	c.Assert(parser.Dump()["tag"], Equals, "myprogram")

}
