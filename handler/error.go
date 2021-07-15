package handler

import "github.com/ONSdigital/log.go/v2/log"

// Error is a structure that contains an error and and a logData map to improve error handling
type Error struct {
	err     error
	logData map[string]interface{}
}

// NewError creates a new Error with the provided error and logData
func NewError(err error, logData map[string]interface{}) *Error {
	return &Error{
		err:     err,
		logData: logData,
	}
}

// Error returns the error string
func (e *Error) Error() string {
	if e.err == nil {
		return "{nil error}"
	}
	return e.err.Error()
}

// logData returns the error logData map
func (e *Error) LogData() map[string]interface{} {
	if e.logData == nil {
		return log.Data{}
	}
	return e.logData
}

// Unwrap returns the error
func (e *Error) Unwrap() error {
	return e.err
}
