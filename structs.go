package syslogparser

import (
	"time"
)

type priority struct {
	p int
	f facility
	s severity
}

type facility struct {
	value int
}

type severity struct {
	value int
}

type rfc3164Parser struct {
	buff    []byte
	cursor  int
	l       int
	header  rfc3164Header
	message rfc3164Message
}

type rfc3164Header struct {
	timestamp time.Time
	hostname  string
}

type rfc3164Message struct {
	tag     string
	content string
}

type logParts map[string]interface{}
