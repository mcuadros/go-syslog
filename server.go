package syslog

import (
	"bufio"
	"errors"
	"net"
	"sync"
	"time"

	"gopkg.in/mcuadros/go-syslog.v2/format"
)

var (
	RFC3164 = &format.RFC3164{} // RFC3164: http://www.ietf.org/rfc/rfc3164.txt
	RFC5424 = &format.RFC5424{} // RFC5424: http://www.ietf.org/rfc/rfc5424.txt
	RFC6587 = &format.RFC6587{} // RFC6587: http://www.ietf.org/rfc/rfc6587.txt
)

const (
	datagramChannelBufferSize = 10
	datagramReadBufferSize    = 1024 * 1024
)

type Server struct {
	listeners               []*net.TCPListener
	connections             []net.Conn
	wait                    sync.WaitGroup
	doneTcp                 chan bool
	doneDatagrams           chan bool
	datagramChannel         chan DatagramMessage
	format                  format.Format
	handler                 Handler
	lastError               error
	readTimeoutMilliseconds int64
}

//NewServer returns a new Server
func NewServer() *Server {
	return &Server{}
}

//Sets the syslog format (RFC3164 or RFC5424 or RFC6587)
func (s *Server) SetFormat(f format.Format) {
	s.format = f
}

//Sets the handler, this handler with receive every syslog entry
func (s *Server) SetHandler(handler Handler) {
	s.handler = handler
}

//Sets the connection timeout for TCP connections, in milliseconds
func (s *Server) SetTimeout(millseconds int64) {
	s.readTimeoutMilliseconds = millseconds
}

//Configure the server for listen on an UDP addr
func (s *Server) ListenUDP(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	connection, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	connection.SetReadBuffer(datagramReadBufferSize)

	s.connections = append(s.connections, connection)
	return nil
}

//Configure the server for listen on an unix socket
func (s *Server) ListenUnixgram(addr string) error {
	unixAddr, err := net.ResolveUnixAddr("unixgram", addr)
	if err != nil {
		return err
	}

	connection, err := net.ListenUnixgram("unixgram", unixAddr)
	if err != nil {
		return err
	}
	connection.SetReadBuffer(datagramReadBufferSize)

	s.connections = append(s.connections, connection)
	return nil
}

//Configure the server for listen on a TCP addr
func (s *Server) ListenTCP(addr string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	s.doneTcp = make(chan bool)
	s.listeners = append(s.listeners, listener)
	return nil
}

//Starts the server, all the go routines goes to live
func (s *Server) Boot() error {
	if s.format == nil {
		return errors.New("please set a valid format")
	}

	if s.handler == nil {
		return errors.New("please set a valid handler")
	}

	for _, listener := range s.listeners {
		s.goAcceptConnection(listener)
	}

	if len(s.connections) > 0 {
		s.goParseDatagrams()
	}

	for _, connection := range s.connections {
		s.goReceiveDatagrams(connection)
	}

	return nil
}

func (s *Server) goAcceptConnection(listener *net.TCPListener) {
	s.wait.Add(1)
	go func(listener *net.TCPListener) {
	loop:
		for {
			select {
			case <-s.doneTcp:
				break loop
			default:
			}
			connection, err := listener.Accept()
			if err != nil {
				continue
			}

			s.goScanConnection(connection)
		}

		s.wait.Done()
	}(listener)
}

func (s *Server) goScanConnection(connection net.Conn) {
	scanner := bufio.NewScanner(connection)
	if sf := s.format.GetSplitFunc(); sf != nil {
		scanner.Split(sf)
	}

	var scanCloser *ScanCloser
	scanCloser = &ScanCloser{scanner, connection}

	s.wait.Add(1)
	go s.scan(scanCloser)
}

func (s *Server) scan(scanCloser *ScanCloser) {
loop:
	for {
		select {
		case <-s.doneTcp:
			break loop
		default:
		}
		if s.readTimeoutMilliseconds > 0 {
			scanCloser.closer.SetReadDeadline(time.Now().Add(time.Duration(s.readTimeoutMilliseconds) * time.Millisecond))
		}
		if scanCloser.Scan() {
			s.parser([]byte(scanCloser.Text()))
		} else {
			break loop
		}
	}
	scanCloser.closer.Close()

	s.wait.Done()
}

func (s *Server) parser(line []byte) {
	parser := s.format.GetParser(line)
	err := parser.Parse()
	if err != nil {
		s.lastError = err
	}

	go s.handler.Handle(parser.Dump(), int64(len(line)), err)
}

//Returns the last error
func (s *Server) GetLastError() error {
	return s.lastError
}

//Kill the server
func (s *Server) Kill() error {
	for _, connection := range s.connections {
		err := connection.Close()
		if err != nil {
			return err
		}
	}

	for _, listener := range s.listeners {
		err := listener.Close()
		if err != nil {
			return err
		}
	}
	// Only need to close channel once to broadcast to all waiting
	if s.doneTcp != nil {
		close(s.doneTcp)
	}
	if s.doneDatagrams != nil {
		close(s.doneDatagrams)
	}
	if s.datagramChannel != nil {
		close(s.datagramChannel)
	}
	return nil
}

//Waits until the server stops
func (s *Server) Wait() {
	s.wait.Wait()
}

type TimeoutCloser interface {
	Close() error
	SetReadDeadline(t time.Time) error
}

type ScanCloser struct {
	*bufio.Scanner
	closer TimeoutCloser
}

type DatagramMessage struct {
	message []byte
	client  string
}

func (s *Server) goReceiveDatagrams(connection net.Conn) {
	s.wait.Add(1)
	go func() {
		packetconn, ok := connection.(net.PacketConn)
		if ok {
		loop:
			for {
				select {
				case <-s.doneDatagrams:
					break loop
				default:
				}
				if s.readTimeoutMilliseconds > 0 {
					connection.SetReadDeadline(time.Now().Add(time.Duration(s.readTimeoutMilliseconds) * time.Millisecond))
				}
				buf := make([]byte, 65536)
				n, addr, err := packetconn.ReadFrom(buf)
				if err == nil {
					// Ignore trailing control characters and NULs
					for ; (n > 0) && (buf[n-1] < 32); n-- {
					}
					if n > 0 {
						s.datagramChannel <- DatagramMessage{buf[:n], addr.String()}
					}
				}
			}
			s.wait.Done()
		}
	}()
}

func (s *Server) goParseDatagrams() {
	s.doneDatagrams = make(chan bool)
	s.datagramChannel = make(chan DatagramMessage, datagramChannelBufferSize)

	s.wait.Add(1)
	go func() {
	loop:
		for {
			select {
			case <-s.doneDatagrams:
				break loop
			case msg, ok := (<-s.datagramChannel):
				if !ok {
					break loop
				}
				s.parser(msg.message)
			}
		}
		s.wait.Done()
	}()
}
