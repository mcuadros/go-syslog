package syslog

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	. "gopkg.in/check.v1"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

func Test(t *testing.T) { TestingT(t) }

type ServerSuite struct {
}

var _ = Suite(&ServerSuite{})
var exampleSyslog = "<31>Dec 26 05:08:46 hostname tag[296]: content"
var exampleSyslogNoTSTagHost = "<14>INFO     leaving (1) step postscripts"
var exampleRFC5424Syslog = "<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - 'su root' failed for lonvick on /dev/pts/8"

func (s *ServerSuite) TestTailFile(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(RFC3164)
	server.SetHandler(handler)
	server.ListenUDP("0.0.0.0:5141")
	server.ListenTCP("0.0.0.0:5141")

	go func(server *Server) {
		time.Sleep(100 * time.Millisecond)

		serverAddr, _ := net.ResolveUDPAddr("udp", "localhost:5141")
		con, _ := net.DialUDP("udp", nil, serverAddr)
		con.Write([]byte(exampleSyslog))
		time.Sleep(100 * time.Millisecond)

		server.Kill()
	}(server)

	server.Boot()
	server.Wait()

	c.Check(handler.LastLogParts["hostname"], Equals, "hostname")
	c.Check(handler.LastLogParts["tag"], Equals, "tag")
	c.Check(handler.LastLogParts["content"], Equals, "content")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslog)))
	c.Check(handler.LastError, IsNil)
}

type HandlerMock struct {
	LastLogParts      format.LogParts
	LastMessageLength int64
	LastError         error
}

func (s *HandlerMock) Handle(logParts format.LogParts, msgLen int64, err error) {
	s.LastLogParts = logParts
	s.LastMessageLength = msgLen
	s.LastError = err
}

type ConnMock struct {
	ReadData       []byte
	ReturnTimeout  bool
	isClosed       bool
	isReadDeadline bool
}

func (c *ConnMock) Read(b []byte) (n int, err error) {
	if c.ReturnTimeout {
		return 0, net.UnknownNetworkError("i/o timeout")
	}
	if c.ReadData != nil {
		l := copy(b, c.ReadData)
		c.ReadData = nil
		return l, nil
	}
	return 0, io.EOF
}

func (c *ConnMock) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (c *ConnMock) Close() error {
	c.isClosed = true
	return nil
}

func (c *ConnMock) LocalAddr() net.Addr {
	return nil
}

func (c *ConnMock) RemoteAddr() net.Addr {
	return nil
}

func (c *ConnMock) SetDeadline(t time.Time) error {
	return nil
}

func (c *ConnMock) SetReadDeadline(t time.Time) error {
	c.isReadDeadline = true
	return nil
}

func (c *ConnMock) SetWriteDeadline(t time.Time) error {
	return nil
}

func (s *ServerSuite) TestConnectionClose(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(RFC3164)
	server.SetHandler(handler)
	con := ConnMock{ReadData: []byte(exampleSyslog)}
	server.goScanConnection(&con)
	server.Wait()
	c.Check(con.isClosed, Equals, true)
}

func (s *ServerSuite) TestConnectionUDPKill(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(RFC5424)
	server.SetHandler(handler)
	con := ConnMock{ReadData: []byte(exampleSyslog)}
	server.goScanConnection(&con)
	server.Kill()
	server.Wait()
	c.Check(con.isClosed, Equals, true)
}

func (s *ServerSuite) TestTcpTimeout(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(RFC3164)
	server.SetHandler(handler)
	server.SetTimeout(10)
	con := ConnMock{ReadData: []byte(exampleSyslog), ReturnTimeout: true}
	c.Check(con.isReadDeadline, Equals, false)
	server.goScanConnection(&con)
	server.Wait()
	c.Check(con.isReadDeadline, Equals, true)
	c.Check(handler.LastLogParts, IsNil)
	c.Check(handler.LastMessageLength, Equals, int64(0))
	c.Check(handler.LastError, IsNil)
}

func (s *ServerSuite) TestUDP3164(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(RFC3164)
	server.SetHandler(handler)
	server.SetTimeout(10)
	server.goParseDatagrams()
	server.datagramChannel <- DatagramMessage{[]byte(exampleSyslog), "0.0.0.0"}
	close(server.datagramChannel)
	server.Wait()
	c.Check(handler.LastLogParts["hostname"], Equals, "hostname")
	c.Check(handler.LastLogParts["tag"], Equals, "tag")
	c.Check(handler.LastLogParts["content"], Equals, "content")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslog)))
	c.Check(handler.LastError, IsNil)
}

func (s *ServerSuite) TestUDP3164NoTag(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(RFC3164)
	server.SetHandler(handler)
	server.SetTimeout(10)
	server.goParseDatagrams()
	server.datagramChannel <- DatagramMessage{[]byte(exampleSyslogNoTSTagHost), "127.0.0.1:45789"}
	close(server.datagramChannel)
	server.Wait()
	c.Check(handler.LastLogParts["hostname"], Equals, "127.0.0.1")
	c.Check(handler.LastLogParts["tag"], Equals, "")
	c.Check(handler.LastLogParts["content"], Equals, "INFO     leaving (1) step postscripts")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslogNoTSTagHost)))
	c.Check(handler.LastError, IsNil)
}

func (s *ServerSuite) TestUDP6587(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(RFC6587)
	server.SetHandler(handler)
	server.SetTimeout(10)
	server.goParseDatagrams()
	framedSyslog := []byte(fmt.Sprintf("%d %s", len(exampleRFC5424Syslog), exampleRFC5424Syslog))
	server.datagramChannel <- DatagramMessage{[]byte(framedSyslog), "0.0.0.0"}
	close(server.datagramChannel)
	server.Wait()
	c.Check(handler.LastLogParts["hostname"], Equals, "mymachine.example.com")
	c.Check(handler.LastLogParts["facility"], Equals, 4)
	c.Check(handler.LastLogParts["message"], Equals, "'su root' failed for lonvick on /dev/pts/8")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleRFC5424Syslog)))
	c.Check(handler.LastError, IsNil)
}

func (s *ServerSuite) TestUDPAutomatic3164(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(Automatic)
	server.SetHandler(handler)
	server.SetTimeout(10)
	server.goParseDatagrams()
	server.datagramChannel <- DatagramMessage{[]byte(exampleSyslog), "0.0.0.0"}
	close(server.datagramChannel)
	server.Wait()
	c.Check(handler.LastLogParts["hostname"], Equals, "hostname")
	c.Check(handler.LastLogParts["tag"], Equals, "tag")
	c.Check(handler.LastLogParts["content"], Equals, "content")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslog)))
	c.Check(handler.LastError, IsNil)
}

func (s *ServerSuite) TestUDPAutomatic5424(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(Automatic)
	server.SetHandler(handler)
	server.SetTimeout(10)
	server.goParseDatagrams()
	server.datagramChannel <- DatagramMessage{[]byte(exampleRFC5424Syslog), "0.0.0.0"}
	close(server.datagramChannel)
	server.Wait()
	c.Check(handler.LastLogParts["hostname"], Equals, "mymachine.example.com")
	c.Check(handler.LastLogParts["facility"], Equals, 4)
	c.Check(handler.LastLogParts["message"], Equals, "'su root' failed for lonvick on /dev/pts/8")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleRFC5424Syslog)))
	c.Check(handler.LastError, IsNil)
}

func (s *ServerSuite) TestUDPAutomatic3164Plus6587OctetCount(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(Automatic)
	server.SetHandler(handler)
	server.SetTimeout(10)
	server.goParseDatagrams()
	framedSyslog := []byte(fmt.Sprintf("%d %s", len(exampleSyslog), exampleSyslog))
	server.datagramChannel <- DatagramMessage{[]byte(framedSyslog), "0.0.0.0"}
	close(server.datagramChannel)
	server.Wait()
	c.Check(handler.LastLogParts["hostname"], Equals, "hostname")
	c.Check(handler.LastLogParts["tag"], Equals, "tag")
	c.Check(handler.LastLogParts["content"], Equals, "content")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslog)))
	c.Check(handler.LastError, IsNil)
}

func (s *ServerSuite) TestUDPAutomatic5424Plus6587OctetCount(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(Automatic)
	server.SetHandler(handler)
	server.SetTimeout(10)
	server.goParseDatagrams()
	framedSyslog := []byte(fmt.Sprintf("%d %s", len(exampleRFC5424Syslog), exampleRFC5424Syslog))
	server.datagramChannel <- DatagramMessage{[]byte(framedSyslog), "0.0.0.0"}
	close(server.datagramChannel)
	server.Wait()
	c.Check(handler.LastLogParts["hostname"], Equals, "mymachine.example.com")
	c.Check(handler.LastLogParts["facility"], Equals, 4)
	c.Check(handler.LastLogParts["message"], Equals, "'su root' failed for lonvick on /dev/pts/8")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleRFC5424Syslog)))
	c.Check(handler.LastError, IsNil)
}

type handlerSlow struct {
	*handlerCounter
	contents []string
}

func (s *handlerSlow) Handle(logParts format.LogParts, msgLen int64, err error) {
	if len(s.contents) == 0 {
		time.Sleep(time.Second)
	}
	s.contents = append(s.contents, logParts["content"].(string))
	s.handlerCounter.Handle(logParts, msgLen, err)
}

func (s *ServerSuite) TestUDPRace(c *C) {
	handler := &handlerSlow{handlerCounter: &handlerCounter{expected: 3, done: make(chan struct{})}}
	server := NewServer()
	server.SetFormat(Automatic)
	server.SetHandler(handler)
	server.SetTimeout(10)
	server.ListenUDP("127.0.0.1:0")
	server.Boot()
	conn, err := net.Dial("udp", server.connections[0].LocalAddr().String())
	c.Assert(err, IsNil)
	_, err = conn.Write([]byte(exampleSyslog + "1"))
	c.Assert(err, IsNil)
	_, err = conn.Write([]byte(exampleSyslog + "2"))
	c.Assert(err, IsNil)
	_, err = conn.Write([]byte(exampleSyslog + "3"))
	c.Assert(err, IsNil)
	conn.Close()
	<-handler.done
	c.Check(handler.contents, DeepEquals, []string{"content1", "content2", "content3"})
}

func (s *ServerSuite) TestTCPRace(c *C) {
	handler := &handlerSlow{handlerCounter: &handlerCounter{expected: 3, done: make(chan struct{})}}
	server := NewServer()
	server.SetFormat(Automatic)
	server.SetHandler(handler)
	server.SetTimeout(10)
	server.ListenTCP("127.0.0.1:0")
	server.Boot()
	conn, err := net.Dial("tcp", server.listeners[0].Addr().String())
	c.Assert(err, IsNil)
	_, err = conn.Write([]byte(exampleSyslog + "1\n"))
	c.Assert(err, IsNil)
	_, err = conn.Write([]byte(exampleSyslog + "2\n"))
	c.Assert(err, IsNil)
	_, err = conn.Write([]byte(exampleSyslog + "3\n"))
	c.Assert(err, IsNil)
	conn.Close()
	<-handler.done
	c.Check(handler.contents, DeepEquals, []string{"content1", "content2", "content3"})
}
