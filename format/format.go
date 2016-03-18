package format

import (
	"bufio"

	"github.com/Xiol/syslogparser"
)

type Format interface {
	GetParser([]byte) syslogparser.LogParser
	GetSplitFunc() bufio.SplitFunc
}
