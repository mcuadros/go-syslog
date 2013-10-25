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

type rfc3164Header struct {
	timestamp time.Time
	hostname  string
}

type rfc3164Message struct {
	tag     string
	content string
}

type logParts map[string]interface{}
