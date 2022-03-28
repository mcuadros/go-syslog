package rfc3164_test

import (
	"fmt"

	"gopkg.in/cnaude/go-syslog.v2/internal/syslogparser/rfc3164"
)

func ExampleNewParser() {
	b := "<34>Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8"
	buff := []byte(b)

	p := rfc3164.NewParser(buff)
	err := p.Parse()
	if err != nil {
		panic(err)
	}

	fmt.Println(p.Dump())
}
