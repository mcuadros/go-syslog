package syslog

import (
	"bufio"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/mcuadros/go-syslog/format"
)

var (
	RFC3164 = &format.RFC3164{} // RFC3164: http://www.ietf.org/rfc/rfc3164.txt
	RFC5424 = &format.RFC5424{} // RFC5424: http://www.ietf.org/rfc/rfc5424.txt
	RFC6587 = &format.RFC6587{} // RFC6587: http://www.ietf.org/rfc/rfc6587.txt
)

type Server struct {
	listeners               []*net.TCPListener
	connections             []net.Conn
	wait                    sync.WaitGroup
	doneTcp                 chan bool
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
func (self *Server) SetFormat(f format.Format) {
	self.format = f
}

//Sets the handler, this handler with receive every syslog entry
func (self *Server) SetHandler(handler Handler) {
	self.handler = handler
}

//Sets the connection timeout for TCP connections, in milliseconds
func (self *Server) SetTimeout(millseconds int64) {
	self.readTimeoutMilliseconds = millseconds
}

//Configure the server for listen on an UDP addr
func (self *Server) ListenUDP(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	connection, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	self.connections = append(self.connections, connection)
	return nil
}

//Configure the server for listen on an unix socket
func (self *Server) ListenUnixgram(addr string) error {
	unixAddr, err := net.ResolveUnixAddr("unixgram", addr)
	if err != nil {
		return err
	}

	connection, err := net.ListenUnixgram("unixgram", unixAddr)
	if err != nil {
		return err
	}

	self.connections = append(self.connections, connection)
	return nil
}

//Configure the server for listen on a TCP addr
func (self *Server) ListenTCP(addr string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	self.doneTcp = make(chan bool)
	self.listeners = append(self.listeners, listener)
	return nil
}

//Starts the server, all the go routines goes to live
func (self *Server) Boot() error {
	if self.format == nil {
		return errors.New("please set a valid format")
	}

	if self.handler == nil {
		return errors.New("please set a valid handler")
	}

	for _, listener := range self.listeners {
		self.goAcceptConnection(listener)
	}

	for _, connection := range self.connections {
		self.goScanConnection(connection, false)
	}

	return nil
}

func (self *Server) goAcceptConnection(listener *net.TCPListener) {
	self.wait.Add(1)
	go func(listener *net.TCPListener) {
	loop:
		for {
			select {
			case <-self.doneTcp:
				break loop
			default:
			}
			connection, err := listener.Accept()
			if err != nil {
				continue
			}

			self.goScanConnection(connection, true)
		}

		self.wait.Done()
	}(listener)
}

func (self *Server) goScanConnection(connection net.Conn, needClose bool) {
	scanner := bufio.NewScanner(connection)
	if sf := self.format.GetSplitFunc(); sf != nil {
		scanner.Split(sf)
	}

	var scanCloser *ScanCloser
	if needClose {
		scanCloser = &ScanCloser{scanner, connection}
	} else {
		scanCloser = &ScanCloser{scanner, nil}
	}

	self.wait.Add(1)
	go self.scan(scanCloser)
}

func (self *Server) scan(scanCloser *ScanCloser) {
	if scanCloser.closer == nil {
		// UDP
		for scanCloser.Scan() {
			self.parser([]byte(scanCloser.Text()))
		}
	} else {
		// TCP
	loop:
		for {
			select {
			case <-self.doneTcp:
				break loop
			default:
			}
			if self.readTimeoutMilliseconds > 0 {
				scanCloser.closer.SetReadDeadline(time.Now().Add(time.Duration(self.readTimeoutMilliseconds) * time.Millisecond))
			}
			if scanCloser.Scan() {
				self.parser([]byte(scanCloser.Text()))
			} else {
				break loop
			}
		}
		scanCloser.closer.Close()
	}

	self.wait.Done()
}

func (self *Server) parser(line []byte) {
	parser := self.format.GetParser(line)
	err := parser.Parse()
	if err != nil {
		self.lastError = err
	}

	go self.handler.Handle(parser.Dump(), int64(len(line)), err)
}

//Returns the last error
func (self *Server) GetLastError() error {
	return self.lastError
}

//Kill the server
func (self *Server) Kill() error {
	for _, connection := range self.connections {
		err := connection.Close()
		if err != nil {
			return err
		}
	}

	for _, listener := range self.listeners {
		err := listener.Close()
		if err != nil {
			return err
		}
	}
	// Only need to close channel once to broadcast to all waiting
	close(self.doneTcp)

	return nil
}

//Waits until the server stops
func (self *Server) Wait() {
	self.wait.Wait()
}

type TimeoutCloser interface {
	Close() error
	SetReadDeadline(t time.Time) error
}

type ScanCloser struct {
	*bufio.Scanner
	closer TimeoutCloser
}
