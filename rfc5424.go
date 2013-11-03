// Note to self : never try to code while looking after your kids
// The result might look like this : https://pbs.twimg.com/media/BXqSuYXIEAAscVA.png

package syslogparser

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

const (
	NILVALUE = '-'
)

func newRfc5424Parser(buff []byte, cursor int, l int) *rfc5424Parser {
	return &rfc5424Parser{
		buff:   buff,
		cursor: cursor,
		l:      l,
	}
}

// HEADER = PRI VERSION SP TIMESTAMP SP HOSTNAME SP APP-NAME SP PROCID SP MSGID
// Note : PRI and VERSION are already parsed in common.go
func (p *rfc5424Parser) parseHeader() (rfc5424Header, error) {
	hdr := rfc5424Header{}

	ts, err := p.parseTimestamp()
	if err != nil {
		return hdr, err
	}

	hdr.timestamp = ts
	p.cursor++

	host, err := p.parseHostname()
	if err != nil {
		return hdr, err
	}

	hdr.hostname = host
	p.cursor++

	appName, err := p.parseAppName()
	if err != nil {
		return hdr, err
	}

	hdr.appName = appName
	p.cursor++

	procId, err := p.parseProcId()
	if err != nil {
		return hdr, nil
	}

	hdr.procId = procId
	p.cursor++

	msgId, err := p.parseMsgId()
	if err != nil {
		return hdr, nil
	}

	hdr.msgId = msgId
	p.cursor++

	return hdr, nil
}

// https://tools.ietf.org/html/rfc5424#section-6.2.3
func (p *rfc5424Parser) parseTimestamp() (time.Time, error) {
	var ts time.Time

	if p.buff[p.cursor] == NILVALUE {
		p.cursor++
		return ts, nil
	}

	fd, err := parseFullDate(p.buff, &p.cursor, p.l)
	if err != nil {
		return ts, err
	}

	if p.buff[p.cursor] != 'T' {
		return ts, ErrInvalidTimeFormat
	}

	p.cursor++

	ft, err := parseFullTime(p.buff, &p.cursor, p.l)
	if err != nil {
		return ts, ErrTimestampUnknownFormat
	}

	nSec, err := toNSec(ft.pt.secFrac)
	if err != nil {
		return ts, err
	}

	ts = time.Date(
		fd.year,
		time.Month(fd.month),
		fd.day,
		ft.pt.hour,
		ft.pt.minute,
		ft.pt.seconds,
		nSec,
		ft.loc,
	)

	return ts, nil
}

// HOSTNAME = NILVALUE / 1*255PRINTUSASCII
func (p *rfc5424Parser) parseHostname() (string, error) {
	return parseHostname(p.buff, &p.cursor, p.l)
}

// APP-NAME = NILVALUE / 1*48PRINTUSASCII
func (p *rfc5424Parser) parseAppName() (string, error) {
	return parseUpToLen(p.buff, &p.cursor, p.l, 48, ErrInvalidAppName)
}

// PROCID = NILVALUE / 1*128PRINTUSASCII
func (p *rfc5424Parser) parseProcId() (string, error) {
	return parseUpToLen(p.buff, &p.cursor, p.l, 128, ErrInvalidProcId)
}

// MSGID = NILVALUE / 1*32PRINTUSASCII
func (p *rfc5424Parser) parseMsgId() (string, error) {
	return parseUpToLen(p.buff, &p.cursor, p.l, 32, ErrInvalidMsgId)
}

// ----------------------------------------------
// https://tools.ietf.org/html/rfc5424#section-6
// ----------------------------------------------

// XXX : bind them to rfc5424Parser ?

// FULL-DATE : DATE-FULLYEAR "-" DATE-MONTH "-" DATE-MDAY
func parseFullDate(buff []byte, cursor *int, l int) (rfc5424FullDate, error) {
	var fd rfc5424FullDate

	year, err := parseYear(buff, cursor, l)
	if err != nil {
		return fd, err
	}

	if buff[*cursor] != '-' {
		return fd, ErrTimestampUnknownFormat
	}

	*cursor++

	month, err := parseMonth(buff, cursor, l)
	if err != nil {
		return fd, err
	}

	if buff[*cursor] != '-' {
		return fd, ErrTimestampUnknownFormat
	}

	*cursor++

	day, err := parseDay(buff, cursor, l)
	if err != nil {
		return fd, err
	}

	fd = rfc5424FullDate{
		year:  year,
		month: month,
		day:   day,
	}

	return fd, nil
}

// DATE-FULLYEAR   = 4DIGIT
func parseYear(buff []byte, cursor *int, l int) (int, error) {
	yearLen := 4

	if *cursor+yearLen > l {
		return 0, ErrEOL
	}

	// XXX : we do not check for a valid year (ie. 1999, 2013 etc)
	// XXX : we only checks the format is correct
	sub := string(buff[*cursor : *cursor+yearLen])

	*cursor += yearLen

	year, err := strconv.Atoi(sub)
	if err != nil {
		return 0, ErrYearInvalid
	}

	return year, nil
}

// DATE-MONTH = 2DIGIT  ; 01-12
func parseMonth(buff []byte, cursor *int, l int) (int, error) {
	return parse2Digits(buff, cursor, l, 1, 12, ErrMonthInvalid)
}

// DATE-MDAY = 2DIGIT  ; 01-28, 01-29, 01-30, 01-31 based on month/year
func parseDay(buff []byte, cursor *int, l int) (int, error) {
	// XXX : this is a relaxed constraint
	// XXX : we do not check if valid regarding February or leap years
	// XXX : we only checks that day is in range [01 -> 31]
	// XXX : in other words this function will not rant if you provide Feb 31th
	return parse2Digits(buff, cursor, l, 1, 31, ErrDayInvalid)
}

// FULL-TIME = PARTIAL-TIME TIME-OFFSET
func parseFullTime(buff []byte, cursor *int, l int) (rfc5424FullTime, error) {
	var loc = new(time.Location)
	var ft rfc5424FullTime

	pt, err := parsePartialTime(buff, cursor, l)
	if err != nil {
		return ft, err
	}

	loc, err = parseTimeOffset(buff, cursor, l)
	if err != nil {
		return ft, err
	}

	ft = rfc5424FullTime{
		pt:  pt,
		loc: loc,
	}

	return ft, nil
}

// PARTIAL-TIME = TIME-HOUR ":" TIME-MINUTE ":" TIME-SECOND[TIME-SECFRAC]
func parsePartialTime(buff []byte, cursor *int, l int) (rfc5424PartialTime, error) {
	var pt rfc5424PartialTime

	hour, minute, err := getHourMinute(buff, cursor, l)
	if err != nil {
		return pt, err
	}

	if buff[*cursor] != ':' {
		return pt, ErrInvalidTimeFormat
	}

	*cursor++

	// ----

	seconds, err := parseSecond(buff, cursor, l)
	if err != nil {
		return pt, err
	}

	pt = rfc5424PartialTime{
		hour:    hour,
		minute:  minute,
		seconds: seconds,
	}

	// ----

	if buff[*cursor] != '.' {
		return pt, nil
	}

	*cursor++

	secFrac, err := parseSecFrac(buff, cursor, l)
	if err != nil {
		return pt, nil
	}
	pt.secFrac = secFrac

	return pt, nil
}

// TIME-HOUR = 2DIGIT  ; 00-23
func parseHour(buff []byte, cursor *int, l int) (int, error) {
	return parse2Digits(buff, cursor, l, 0, 23, ErrHourInvalid)
}

// TIME-MINUTE = 2DIGIT  ; 00-59
func parseMinute(buff []byte, cursor *int, l int) (int, error) {
	return parse2Digits(buff, cursor, l, 0, 59, ErrMinuteInvalid)
}

// TIME-SECOND = 2DIGIT  ; 00-59
func parseSecond(buff []byte, cursor *int, l int) (int, error) {
	return parse2Digits(buff, cursor, l, 0, 59, ErrSecondInvalid)
}

// TIME-SECFRAC = "." 1*6DIGIT
func parseSecFrac(buff []byte, cursor *int, l int) (float64, error) {
	maxDigitLen := 6

	max := *cursor + maxDigitLen
	from := *cursor
	to := from

	for to = from; to < max; to++ {
		if to >= l {
			break
		}

		c := buff[to]
		if !isDigit(c) {
			break
		}
	}

	sub := string(buff[from:to])
	if len(sub) == 0 {
		return 0, ErrSecFracInvalid
	}

	secFrac, err := strconv.ParseFloat("0."+sub, 64)
	*cursor = to
	if err != nil {
		return 0, ErrSecFracInvalid
	}

	return secFrac, nil
}

// TIME-OFFSET = "Z" / TIME-NUMOFFSET
func parseTimeOffset(buff []byte, cursor *int, l int) (*time.Location, error) {

	if buff[*cursor] == 'Z' {
		*cursor++
		return time.UTC, nil
	}

	return parseNumericalTimeOffset(buff, cursor, l)
}

// TIME-NUMOFFSET  = ("+" / "-") TIME-HOUR ":" TIME-MINUTE
func parseNumericalTimeOffset(buff []byte, cursor *int, l int) (*time.Location, error) {
	var loc = new(time.Location)

	sign := buff[*cursor]

	if (sign != '+') && (sign != '-') {
		return loc, ErrTimeZoneInvalid
	}

	*cursor++

	hour, minute, err := getHourMinute(buff, cursor, l)
	if err != nil {
		return loc, err
	}

	tzStr := fmt.Sprintf("%s%02d:%02d", string(sign), hour, minute)
	tmpTs, err := time.Parse("-07:00", tzStr)
	if err != nil {
		return loc, err
	}

	return tmpTs.Location(), nil
}

func getHourMinute(buff []byte, cursor *int, l int) (int, int, error) {
	hour, err := parseHour(buff, cursor, l)
	if err != nil {
		return 0, 0, err
	}

	if buff[*cursor] != ':' {
		return 0, 0, ErrInvalidTimeFormat
	}

	*cursor++

	minute, err := parseMinute(buff, cursor, l)
	if err != nil {
		return 0, 0, err
	}

	return hour, minute, nil
}

func toNSec(sec float64) (int, error) {
	_, frac := math.Modf(sec)
	fracStr := strconv.FormatFloat(frac, 'f', 9, 64)
	fracInt, err := strconv.Atoi(fracStr[2:])
	if err != nil {
		return 0, err
	}

	return fracInt, nil
}

// ------------------------------------------------
// https://tools.ietf.org/html/rfc5424#section-6.3
// ------------------------------------------------

func parseStructuredData(buff []byte, cursor *int, l int) (string, error) {
	var sdData string
	var found bool

	if buff[*cursor] == NILVALUE {
		*cursor++
		return sdData, nil
	}

	if buff[*cursor] != '[' {
		return sdData, ErrNoStructuredData
	}

	from := *cursor
	to := from

	for to = from; to < l; to++ {
		if found {
			break
		}

		b := buff[to]

		if b == ']' {
			switch t := to + 1; {
			case t == l:
				found = true
			case t <= l && buff[t] == ' ':
				found = true
			}
		}
	}

	if found {
		*cursor = to
		return string(buff[from:to]), nil
	}

	return sdData, ErrNoStructuredData
}

func parseUpToLen(buff []byte, cursor *int, l int, maxLen int, e error) (string, error) {
	var to int
	var found bool
	var result string

	max := *cursor + maxLen

	for to = *cursor; (to < max) && (to < l); to++ {
		if buff[to] == ' ' {
			found = true
			break
		}
	}

	if found {
		result = string(buff[*cursor:to])
	}

	*cursor = to

	if found {
		return result, nil
	}

	return "", e
}
