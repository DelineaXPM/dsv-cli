package errors

import (
	"fmt"
	"net/http"
	"strings"
)

// ApiError is an improved error class
type ApiError struct {
	stack        []string
	httpResponse *http.Response
}

func initError() *ApiError {
	return &ApiError{stack: []string{}}
}

// New creates a new error from an error
func New(err error) *ApiError {
	if err == nil || err.Error() == "" {
		return nil
	}
	return &ApiError{
		stack: []string{err.Error()},
	}
}

// NewS creates a new error from a string
func NewS(err string) *ApiError {
	if err == "" {
		return nil
	}
	return &ApiError{
		stack: []string{strings.TrimSpace(err)},
	}
}

// NewF creates a new formatted error
func NewF(format string, a ...interface{}) *ApiError {
	if format == "" {
		return nil
	}
	return &ApiError{
		stack: []string{fmt.Sprintf(format, a...)},
	}
}

func (e *ApiError) Error() string {
	if e == nil {
		return ""
	}
	return strings.Join(e.stack, "\n")
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

// Add adds an error to the current error
func (e *ApiError) Add(err string) {
	if e == nil {
		e = initError()
	}
	if err == "" {
		return
	}
	e.stack = append([]string{err}, e.stack...)
}

// Grow adds to an error and returns it.
func (e *ApiError) Grow(err string) *ApiError {
	e.Add(err)
	return e
}

// GrowIf adds to an error, if it is not nil, and returns it.
func (e *ApiError) GrowIf(err string) *ApiError {
	if e == nil {
		return e
	}
	return e.Grow(err)
}

// Or coalesces two errors.
func (e *ApiError) Or(f *ApiError) *ApiError {
	if e == nil || e.stack == nil || len(e.stack) < 1 {
		return f
	}
	return e
}

// And ands two errors
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

// Convert data and a std err to an api err
func Convert(data []byte, err error) ([]byte, *ApiError) {
	return data, New(err)
}
