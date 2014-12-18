package syslog

import (
	"github.com/jeromer/syslogparser"
	"github.com/jeromer/syslogparser/rfc3164"
	"github.com/jeromer/syslogparser/rfc5424"
)

type Format interface {
	GetParser([]byte) syslogparser.LogParser
}

type format3164 struct {}
type format5424 struct {}

var (
	RFC3164 = &format3164{}
	RFC5424 = &format5424{}
)

func (fmt *format3164) GetParser (line []byte) syslogparser.LogParser {
	return rfc3164.NewParser(line)
}

func (fmt *format5424) GetParser (line []byte) syslogparser.LogParser {
	return rfc5424.NewParser(line)
}
