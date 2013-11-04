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

// ----

type rfc3164Parser struct {
	buff     []byte
	cursor   int
	l        int
	priority priority
	version  int
	header   rfc3164Header
	message  rfc3164Message
}

type rfc3164Header struct {
	timestamp time.Time
	hostname  string
}

type rfc3164Message struct {
	tag     string
	content string
}

// ----

type rfc5424Parser struct {
	buff           []byte
	cursor         int
	l              int
	header         rfc5424Header
	structuredData string
	message        string
}

type rfc5424Header struct {
	priority  priority
	version   int
	timestamp time.Time
	hostname  string
	appName   string
	procId    string
	msgId     string
}

type rfc5424PartialTime struct {
	hour    int
	minute  int
	seconds int
	secFrac float64
}

type rfc5424FullTime struct {
	pt  rfc5424PartialTime
	loc *time.Location
}

type rfc5424FullDate struct {
	year  int
	month int
	day   int
}

type LogParts map[string]interface{}
