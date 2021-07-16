package event

import "github.com/ONSdigital/log.go/v2/log"

// Error is the handler package's error type. Is not meant to be compared as a
// a type, but information should be extracted via the interfaces
// it implements with callback functions. Is not guaranteed to remain exported
// so shouldn't be treated as such.
type Error struct {
	err     error
	logData map[string]interface{}
}

// NewError creates a new Error
func NewError(err error, logData map[string]interface{}) *Error {
	return &Error{
		err:     err,
		logData: logData,
	}
}

// Error implements the Go standard error interface
func (e *Error) Error() string {
	if e.err == nil {
		return "{nil error}"
	}
	return e.err.Error()
}

// LogData implements the DataLogger interface which allows you extract
// embedded log.Data from an error
func (e *Error) LogData() map[string]interface{} {
	if e.logData == nil {
		return log.Data{}
	}
	return e.logData
}

// Unwrap returns the wrapped error
func (e *Error) Unwrap() error {
	return e.err
}
