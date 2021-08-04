package rack

import (
	"errors"
	"net/http"
)

type (
	// StatusError represents a status code error
	StatusError struct {
		code int
		err  error
	}

	statusError interface {
		Code() int
		error
	}
)

// StatusCode returns the status code for the specified error
func StatusCode(err error) int {
	var se statusError
	if errors.As(err, &se) {
		return se.Code()
	}

	return http.StatusInternalServerError
}

// WrapError wraps the specified error
func WrapError(code int, err error) *StatusError {
	return &StatusError{
		code: code,
		err:  err,
	}
}

// Code returns the error status code
func (e *StatusError) Code() int {
	return e.code
}

// Error returns the error message
func (e *StatusError) Error() string {
	return e.err.Error()
}

// Unwrap returns the wrapped error
func (e *StatusError) Unwrap() error {
	return e.err
}
