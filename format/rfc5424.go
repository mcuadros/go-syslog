package format

import (
	"bufio"

	"gopkg.in/cnaude/go-syslog.v2/internal/syslogparser/rfc5424"
)

type RFC5424 struct{}

func (f *RFC5424) GetParser(line []byte) LogParser {
	return &parserWrapper{rfc5424.NewParser(line)}
}

func (f *RFC5424) GetSplitFunc() bufio.SplitFunc {
	return rfc5424ScannerSplit
}

func rfc5424ScannerSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	// return all of the data without splitting
	return len(data), data, nil
}
