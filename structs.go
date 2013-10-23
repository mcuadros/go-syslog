package syslogparser

type Priority struct {
	Facility
	Severity
}

type Facility struct {
	Value int
}

type Severity struct {
	Value int
}
