package syslogparser

import (
	"strconv"
)

const (
	PRI_PART_START = '<'
	PRI_PART_END   = '>'

	VERSION_NONE = -1
)

// https://tools.ietf.org/html/rfc3164#section-4.1
func ParsePriority(buff []byte, start *int, l int) (Priority, error) {
	pri := newPriority(0)

	if l <= 0 {
		return pri, ErrPriorityEmpty
	}

	if buff[*start] != PRI_PART_START {
		return pri, ErrPriorityNoStart
	}

	cursor := 1
	priDigit := 0

	for cursor < l {
		if cursor >= 5 {
			return pri, ErrPriorityTooLong
		}

		c := buff[cursor]

		if c == PRI_PART_END {
			if cursor == 1 {
				return pri, ErrPriorityTooShort
			}

			return newPriority(priDigit), nil
		}

		if isDigit(c) {
			v, e := strconv.Atoi(string(c))
			if e != nil {
				return pri, e
			}

			priDigit = (priDigit * 10) + v
		} else {
			return pri, ErrPriorityNonDigit
		}

		cursor++
	}

	return pri, ErrPriorityNoEnd
}

// https://tools.ietf.org/html/rfc5424#section-6.2.2
func ParseVersion(buff []byte, start *int, l int) (int, error) {
	if *start >= l {
		return VERSION_NONE, ErrVersionNotFound
	}

	c := buff[*start]

	if !isDigit(c) {
		return VERSION_NONE, ErrVersionNonDigit
	}

	v, e := strconv.Atoi(string(c))
	if e != nil {
		return VERSION_NONE, e
	}

	return v, nil
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func newPriority(p int) Priority {
	// The Priority value is calculated by first multiplying the Facility
	// number by 8 and then adding the numerical value of the Severity.

	f := Facility{Value: p / 8}
	s := Severity{Value: p % 8}

	return Priority{
		Facility: f,
		Severity: s,
	}
}
