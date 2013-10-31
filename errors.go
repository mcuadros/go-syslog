package syslogparser

var (
	ErrEOL     = &ParserError{"End of log line"}
	ErrNoSpace = &ParserError{"No space found"}

	ErrYearInvalid       = &ParserError{"Invalid year in timestamp"}
	ErrMonthInvalid      = &ParserError{"Invalid month in timestamp"}
	ErrDayInvalid        = &ParserError{"Invalid day in timestamp"}
	ErrHourInvalid       = &ParserError{"Invalid hour in timestamp"}
	ErrMinuteInvalid     = &ParserError{"Invalid minute in timestamp"}
	ErrSecondInvalid     = &ParserError{"Invalid second in timestamp"}
	ErrSecFracInvalid    = &ParserError{"Invalid fraction of second in timestamp"}
	ErrTimeZoneInvalid   = &ParserError{"Invalid time zone in timestamp"}
	ErrInvalidTimeFormat = &ParserError{"Invalid time format"}

	ErrInvalidAppName = &ParserError{"Invalid app name"}

	ErrInvalidProcId = &ParserError{"Invalid proc ID"}
	ErrInvalidMsgId  = &ParserError{"Invalid msg ID"}

	ErrPriorityNoStart  = &ParserError{"No start char found for priority"}
	ErrPriorityEmpty    = &ParserError{"Priority field empty"}
	ErrPriorityNoEnd    = &ParserError{"No end char found for priority"}
	ErrPriorityTooShort = &ParserError{"Priority field too short"}
	ErrPriorityTooLong  = &ParserError{"Priority field too long"}
	ErrPriorityNonDigit = &ParserError{"Non digit found in priority"}

	ErrVersionNotFound = &ParserError{"Can not find version"}

	ErrTimestampUnknownFormat = &ParserError{"Timestamp format unknown"}

	ErrTagTooLong = &ParserError{"Tag name too long"}
)

type ParserError struct {
	ErrorString string
}

func (err *ParserError) Error() string {
	return err.ErrorString
}
