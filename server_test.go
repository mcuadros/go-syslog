package syslog

import (
	"net"
	"testing"
	"time"
)

import . "launchpad.net/gocheck"
import "github.com/jeromer/syslogparser"

func Test(t *testing.T) { TestingT(t) }

type ServerSuite struct {
}

var _ = Suite(&ServerSuite{})
var exampleSyslog = "<31>Dec 26 05:08:46 hostname tag[296]: content"

func (s *ServerSuite) TestTailFile(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(RFC3164)
	server.SetHandler(handler)
	server.ListenUDP("0.0.0.0:5141")

	go func(server *Server) {
		time.Sleep(100 * time.Microsecond)

		serverAddr, _ := net.ResolveUDPAddr("udp", "localhost:5141")
		con, _ := net.DialUDP("udp", nil, serverAddr)
		con.Write([]byte(exampleSyslog))
		time.Sleep(100 * time.Microsecond)

		server.Kill()
	}(server)

	server.Boot()
	server.Wait()

	c.Check(handler.LastLogParts["hostname"], Equals, "hostname")
	c.Check(handler.LastLogParts["tag"], Equals, "tag")
	c.Check(handler.LastLogParts["content"], Equals, "content")
}

type HandlerMock struct {
	LastLogParts syslogparser.LogParts
}

func (self *HandlerMock) Handle(logParts syslogparser.LogParts) {
	self.LastLogParts = logParts
}
