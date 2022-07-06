package errors

import (
	"fmt"
	"net/http"
	"strings"
)

// ApiError is an improved error class.
type ApiError struct {
	stack        []string
	httpResponse *http.Response
}

// New creates a new error from an error.
func New(err error) *ApiError {
	if err == nil {
		return nil
	}
	return NewS(err.Error())
}

// NewF creates a new formatted error.
func NewF(format string, a ...interface{}) *ApiError {
	if format == "" {
		return nil
	}
	return NewS(fmt.Sprintf(format, a...))
}

// NewS creates a new error from a string.
func NewS(msg string) *ApiError {
	if msg == "" {
		return nil
	}
	return &ApiError{
		stack: []string{strings.TrimSpace(msg)},
	}
}

func (e *ApiError) Error() string {
	return e.String()
}

func (e *ApiError) String() string {
	if e == nil {
		return ""
	}
	return strings.Join(e.stack, "\n")
}

func (e *ApiError) WithResponse(httpResponse *http.Response) *ApiError {
	e.httpResponse = httpResponse
	return e
}

func (e *ApiError) HttpResponse() *http.Response {
	return e.httpResponse
}

// Grow adds to an error and returns it.
func (e *ApiError) Grow(msg string) *ApiError {
	if e == nil {
		return nil
	}
	if msg == "" {
		return e
	}
	e.stack = append([]string{msg}, e.stack...)
	return e
}

// Or coalesces two errors.
func (e *ApiError) Or(f *ApiError) *ApiError {
	if e == nil || e.stack == nil || len(e.stack) < 1 {
		return f
	}
	return e
}

// And ands two errors.
func (e *ApiError) And(f *ApiError) *ApiError {
	if f == nil || f.stack == nil || len(f.stack) < 1 {
		return e
	}
	if e == nil || e.stack == nil || len(e.stack) < 1 {
		return f
	}
	e.stack = append(e.stack, "and")
	e.stack = append(e.stack, f.stack...)
	return e
}

// Convert data and a std err to an api err.
func Convert(data []byte, err error) ([]byte, *ApiError) {
	return data, New(err)
}
