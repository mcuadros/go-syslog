package syslogparser

var (
	ErrPriorityNoStart  = &ParserError{"No start char found for priority"}
	ErrPriorityEmpty    = &ParserError{"Priority field empty"}
	ErrPriorityNoEnd    = &ParserError{"No end char found for priority"}
	ErrPriorityTooShort = &ParserError{"Priority field too short"}
	ErrPriorityTooLong  = &ParserError{"Priority field too long"}
	ErrPriorityNonDigit = &ParserError{"Non digit found in priority"}

	ErrVersionNotFound = &ParserError{"Can not find version"}
	ErrVersionNonDigit = &ParserError{"Non digit found for version"}
)

type ParserError struct {
	ErrorString string
}

func (err *ParserError) Error() string {
	return err.ErrorString
}
