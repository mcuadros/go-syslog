package syslog

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestSingleSplit(t *testing.T) {
	find := "I am test."
	buf := strings.NewReader("10 " + find)
	scanner := bufio.NewScanner(buf)
	scanner.Split(rfc6587ScannerSplit)
	if r := scanner.Scan(); !r {
		t.Error("Expected Scan() to return true, but didn't")
	}
	if found := scanner.Text(); found != find {
		t.Errorf("Expected the right ('%s') token, but got: '%s'\n", find, found)
	}
}

func TestMultiSplit(t *testing.T) {
	find := []string{
		"I am test.",
		"I am test 2.",
		"hahahahah",
	}
	buf := new(bytes.Buffer)
	for _, i := range find {
		fmt.Fprintf(buf, "%d %s", len(i), i)
	}
	scanner := bufio.NewScanner(buf)
	scanner.Split(rfc6587ScannerSplit)

	i := 0
	for scanner.Scan() {
		i++
	}

	if i != len(find) {
		t.Errorf("Expected to find %d items, but found: %d\n", len(find), i)
	}
}

func TestBadSplit(t *testing.T) {
	find := "I am test.2 ab"
	buf := strings.NewReader("9 " + find)
	scanner := bufio.NewScanner(buf)
	scanner.Split(rfc6587ScannerSplit)
	if r := scanner.Scan(); !r {
		t.Error("Expected Scan() to return true, but didn't")
	}
	if found := scanner.Text(); found != find[0:9] {
		t.Errorf("Expected to find %s, but found %s.", find[0:9], found)
	}
	if r := scanner.Scan(); r {
		t.Error("Expected Scan() to return false, but didn't")
	}
	if err := scanner.Err(); err == nil {
		t.Error("Expected an error, but didn't get one")
	} else {
		t.Log("Error was: ", err)
	}

}
