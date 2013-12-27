package syslog

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"sync"
)

import "github.com/jeromer/syslogparser"
import "github.com/jeromer/syslogparser/rfc3164"
import "github.com/jeromer/syslogparser/rfc5424"

type Format int

const (
	RFC3164           Format = 1 + iota // RFC3164: http://www.ietf.org/rfc/rfc3164.txt
	RFC3164_NO_STRICT                   // RFC3164: but allows tags longer than 32 chars
	RFC5423                             // RFC5423: http://www.ietf.org/rfc/rfc5424.txt
)

type Server struct {
	scanners    []*bufio.Scanner
	listeners   []*net.TCPListener
	connections []net.Conn
	wait        sync.WaitGroup
	format      Format
	handler     Handler
	lastError   error
}

//NewServer returns a new Server
func NewServer() *Server {
	server := new(Server)

	return server
}

//Sets the syslog format (RFC3164 or RFC5423)
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

	for _, listerner := range self.listeners {
		self.goAcceptConnection(listerner)
	}

	for _, connection := range self.connections {
		self.goScanConnection(connection)
	}

	return nil
}

func (self *Server) goAcceptConnection(listerner *net.TCPListener) {
	self.wait.Add(1)
	go func(listerner *net.TCPListener) {
		for {
			connection, err := listerner.Accept()
			if err != nil {
				continue
			}

			self.goScanConnection(connection)
		}

		self.wait.Done()
	}(listerner)
}

func (self *Server) goScanConnection(connection net.Conn) {
	scanner := bufio.NewScanner(connection)
	self.scanners = append(self.scanners, scanner)

	self.wait.Add(1)
	go self.scan(scanner)
}

func (self *Server) scan(scanner *bufio.Scanner) {
	for scanner.Scan() {
		self.parser([]byte(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("reading standard input:", err)
	}

	self.wait.Done()
}

func (self *Server) parser(line []byte) {
	var parser syslogparser.LogParser

	switch self.format {
	case RFC3164:
		parser = self.getParserRFC3164(line, true)
	case RFC3164_NO_STRICT:
		parser = self.getParserRFC3164(line, false)
	case RFC5423:
		parser = self.getParserRFC5424(line)
	}

	if err := parser.Parse(); err != nil {
		self.lastError = err
	}

	go self.handler.Handle(parser.Dump())
}

func (self *Server) getParserRFC3164(line []byte, strict bool) *rfc3164.Parser {
	parser := rfc3164.NewParser(line)
	parser.SetStrict(strict)

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

//Waits until the server stops
func (self *Server) Wait() {
	self.wait.Wait()
}
