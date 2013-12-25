package syslog

import (
	"fmt"
	"net"
	"testing"
	"time"
)

import . "launchpad.net/gocheck"

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
	err := server.ListenUDP("localhost:5142")
	fmt.Println(err)

	server.Boot()
	server.Wait()
}
