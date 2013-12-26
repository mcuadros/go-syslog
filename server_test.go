package syslog

import (
	"fmt"
	"net"
	"testing"
	"time"
)

import . "launchpad.net/gocheck"
import "github.com/jeromer/syslogparser"

func Test(t *testing.T) { TestingT(t) }

type ServerSuite struct{}

var _ = Suite(&ServerSuite{})

func (s *ServerSuite) TestTailFile(c *C) {
	go func() {
		time.Sleep(100 * time.Microsecond)
		serverAddr, _ := net.ResolveUDPAddr("udp", "localhost:5142")
		con, _ := net.DialUDP("udp", nil, serverAddr)
		con.Write([]byte("foo\n"))
	}()

	server := NewServer()
	server.SetFormat(RFC3164_NO_STRICT)
	server.SetHandler(new(HandlerMock))

	err := server.ListenUDP("0.0.0.0:514")
	fmt.Println(err)

	server.Boot()
	server.Wait()
}

type HandlerMock struct {
}

func (self *HandlerMock) Handle(logParts syslogparser.LogParts) {
	fmt.Println(logParts)
}
