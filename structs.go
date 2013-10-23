package syslogparser

type priority struct {
	f facility
	s severity
}

type facility struct {
	value int
}

type severity struct {
	value int
}
