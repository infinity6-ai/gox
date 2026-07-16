package errorz

import (
	"github.com/infinity6-ai/gox/commonz/runtimez"
)

// customError holds the original error and the stack trace data.
type customError struct {
	cause error
	stack string
}

// Error implements the error interface.
// It ONLY returns the original error message, keeping logs clean.
func (e *customError) Error() string {
	return e.cause.Error()
}

// Unwrap allows errors.Is and errors.As to work perfectly.
func (e *customError) Unwrap() error {
	return e.cause
}

// StackTrace exposes the stack trace to loggers that know to look for it.
func (e *customError) StackTrace() string {
	return e.stack
}

// DetailedError is an error with a stack trace.
type DetailedError interface {
	error
	Unwrap() error
	StackTrace() string
}

// New wraps the error and captures the stack.
func New(cause error) DetailedError {
	if cause == nil {
		return nil
	}

	return &customError{
		cause: cause,
		stack: runtimez.StackTraceString(3),
	}
}
