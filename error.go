package rack

import (
	"errors"
	"net/http"
)

// StatusError represents a status code error
type StatusError struct {
	code    int
	message string
	inner   error
}

// StatusCode returns the status code for the specified error
func StatusCode(err error) int {
	se := new(StatusError)
	if errors.As(err, &se) {
		return se.Code()
	}

	return http.StatusInternalServerError
}

// NewError returns a new status error
func NewError(code int, message string) *StatusError {
	return &StatusError{
		code:    code,
		message: message,
	}
}

// WrapError wraps the specified error
func WrapError(code int, err error) *StatusError {
	return &StatusError{
		code:    code,
		message: err.Error(),
		inner:   err,
	}
}

// Code returns the error status code
func (e *StatusError) Code() int {
	return e.code
}

// Error returns the error message
func (e *StatusError) Error() string {
	return e.message
}

// Unwrap returns the wrapped error
func (e *StatusError) Unwrap() error {
	return e.inner
}
