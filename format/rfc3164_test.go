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

	find := `<13>May  1 20:51:40 myhostname myprogram: ciao`
	parser := f.GetParser([]byte(find))
	err := parser.Parse()
	c.Assert(err, IsNil)
	c.Assert(parser.Dump()["content"], Equals, "ciao")
	c.Assert(parser.Dump()["hostname"], Equals, "myhostname")
	c.Assert(parser.Dump()["tag"], Equals, "myprogram")

}
func (s *FormatSuite) TestRFC3164_CorrectParsingTypicalWithPID(c *C) {
	f := RFC3164{}

	find := `<13>May  1 20:51:40 myhostname myprogram[42]: ciao`
	parser := f.GetParser([]byte(find))
	err := parser.Parse()
	c.Assert(err, IsNil)
	c.Assert(parser.Dump()["content"], Equals, "ciao")
	c.Assert(parser.Dump()["hostname"], Equals, "myhostname")
	c.Assert(parser.Dump()["tag"], Equals, "myprogram")

}

func (s *FormatSuite) TestRFC3164_CorrectParsingGNU(c *C) {
	// GNU implementation of syslog() has a variant: hostname is missing
	f := RFC3164{}

	find := `<13>May  1 20:51:40 myprogram: ciao`
	parser := f.GetParser([]byte(find))
	err := parser.Parse()
	c.Assert(err, IsNil)
	c.Assert(parser.Dump()["content"], Equals, "ciao")
	// c.Assert(parser.Dump()["hostname"], Equals, "myhostname")
	c.Assert(parser.Dump()["tag"], Equals, "myprogram")

}

func (s *FormatSuite) TestRFC3164_CorrectParsingJournald(c *C) {
	// GNU implementation of syslog() has a variant: hostname is missing
	// systemd uses it, and typically also passes PID
	f := RFC3164{}

	find := `<78>May  1 20:51:02 myprog[153]: blah`
	parser := f.GetParser([]byte(find))
	err := parser.Parse()
	c.Assert(err, IsNil)
	c.Assert(parser.Dump()["content"], Equals, "blah")
	// c.Assert(parser.Dump()["hostname"], Equals, "myhostname")
	c.Assert(parser.Dump()["tag"], Equals, "myprog")

}
