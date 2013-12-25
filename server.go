package syslog

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type Server struct {
	scanners    []*bufio.Scanner
	connections []net.Conn
	wait        sync.WaitGroup
}

func NewServer() *Server {
	server := new(Server)

	return server
}

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

func (self *Server) Boot() error {
	for _, connection := range self.connections {
		scanner := self.createScannerFromConnection(connection)
		self.scanners = append(self.scanners, scanner)

		self.wait.Add(1)
		go self.scan(scanner)
	}

	return nil
}

func (self *Server) createScannerFromConnection(connection net.Conn) *bufio.Scanner {
	return bufio.NewScanner(connection)
}

func (self *Server) scan(scanner *bufio.Scanner) {
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("reading standard input:", err)
	}

	self.wait.Done()
}

func (self *Server) Wait() {
	self.wait.Wait()
}
