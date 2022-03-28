package format

import (
	"bufio"

	"github.com/cnaude/go-syslog/internal/syslogparser/rfc3164"
)

type RFC3164 struct{}

func (f *RFC3164) GetSplitFunc() bufio.SplitFunc {
	return rfc3164ScannerSplit
}

func rfc3164ScannerSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	// return all of the data without splitting
	return len(data), data, nil
}
