package syslog

import (
	"bufio"
	"io"
	"net"
	"testing"
	"time"

	"gopkg.in/mcuadros/go-syslog.v2/format"
)

type noopFormatter struct{}

func (noopFormatter) Parse() error {
	return nil
}

func (noopFormatter) Dump() format.LogParts {
	return format.LogParts{}
}

func (noopFormatter) Location(*time.Location) {}

func (n noopFormatter) GetParser(l []byte) format.LogParser {
	return n
}

func (n noopFormatter) GetSplitFunc() bufio.SplitFunc {
	return nil
}

type handlerCounter struct {
	expected int
	current  int
	done     chan struct{}
}

func (s *handlerCounter) Handle(logParts format.LogParts, msgLen int64, err error) {
	s.current++
	if s.current == s.expected {
		close(s.done)
	}
}

type fakePacketConn struct {
	*io.PipeReader
}

func (c *fakePacketConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	n, err = c.PipeReader.Read(b)
	return
}
func (c *fakePacketConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	return 0, nil
}
func (c *fakePacketConn) Close() error {
	return nil
}
func (c *fakePacketConn) LocalAddr() net.Addr {
	return nil
}
func (c *fakePacketConn) SetDeadline(t time.Time) error {
	return nil
}
func (c *fakePacketConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (c *fakePacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func BenchmarkDatagramNoFormatting(b *testing.B) {
	handler := &handlerCounter{expected: b.N, done: make(chan struct{})}
	server := NewServer()
	defer server.Kill()
	server.SetFormat(noopFormatter{})
	server.SetHandler(handler)
	reader, writer := io.Pipe()
	server.goReceiveDatagrams(&fakePacketConn{PipeReader: reader})
	server.goParseDatagrams()
	msg := []byte(exampleSyslog + "\n")
	b.SetBytes(int64(len(msg)))
	for i := 0; i < b.N; i++ {
		writer.Write(msg)
	}
	<-handler.done
}

func BenchmarkTCPNoFormatting(b *testing.B) {
	handler := &handlerCounter{expected: b.N, done: make(chan struct{})}
	server := NewServer()
	defer server.Kill()
	server.SetFormat(noopFormatter{})
	server.SetHandler(handler)
	server.ListenTCP("127.0.0.1:0")
	server.Boot()
	conn, _ := net.DialTimeout("tcp", server.listeners[0].Addr().String(), time.Second)
	msg := []byte(exampleSyslog + "\n")
	b.SetBytes(int64(len(msg)))
	for i := 0; i < b.N; i++ {
		conn.Write(msg)
	}
	<-handler.done
}
