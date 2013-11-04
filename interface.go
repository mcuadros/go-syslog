package syslogparser

type LogParser interface {
	Parse() error
	Dump() LogParts
}
