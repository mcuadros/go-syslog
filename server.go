package syslog

import (
	"bufio"
	"errors"
	"net"
	"sync"
)

import "github.com/jeromer/syslogparser"
import "github.com/jeromer/syslogparser/rfc3164"
import "github.com/jeromer/syslogparser/rfc5424"

type Format int

const (
	RFC3164 Format = 1 + iota // RFC3164: http://www.ietf.org/rfc/rfc3164.txt
	RFC5424                   // RFC5424: http://www.ietf.org/rfc/rfc5424.txt
)

type Server struct {
	listeners   []*net.TCPListener
	connections []net.Conn
	wait        sync.WaitGroup
	doneTcp		chan bool
	format      Format
	handler     Handler
	lastError   error
}

//NewServer returns a new Server
func NewServer() *Server {
	server := new(Server)

	return server
}

//Sets the syslog format (RFC3164 or RFC5424)
func (self *Server) SetFormat(format Format) {
	self.format = format
}

//Sets the handler, this halder with receive every syslog entry
func (self *Server) SetHandler(handler Handler) {
	self.handler = handler
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
	if self.format == 0 {
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

type Closer interface {
	Close() error
}

type ScanCloser struct {
	*bufio.Scanner
	closer Closer
}

func (self *Server) goScanConnection(connection net.Conn, needClose bool) {
	scanner := bufio.NewScanner(connection)

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
	for scanCloser.Scan() {
		self.parser([]byte(scanCloser.Text()))
	}

	if scanCloser.closer != nil {
		scanCloser.closer.Close()
	}

	self.wait.Done()
}

func (self *Server) parser(line []byte) {
	var parser syslogparser.LogParser

	switch self.format {
	case RFC3164:
		parser = self.getParserRFC3164(line)
	case RFC5424:
		parser = self.getParserRFC5424(line)
	}

	if err := parser.Parse(); err != nil {
		self.lastError = err
	}

	go self.handler.Handle(parser.Dump())
}

func (self *Server) getParserRFC3164(line []byte) *rfc3164.Parser {
	parser := rfc3164.NewParser(line)

	return parser
}

func (self *Server) getParserRFC5424(line []byte) *rfc5424.Parser {
	parser := rfc5424.NewParser(line)

	return parser
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
		close(self.doneTcp)
	}

	return nil
}

//Waits until the server stops
func (self *Server) Wait() {
	self.wait.Wait()
}
