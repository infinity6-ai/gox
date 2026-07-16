package errorz

import (
	"errors"

	"github.com/infinity6-ai/gox/commonz/runtimez"
)

// detailedError holds the original error and the stack trace data.
type detailedError struct {
	cause error
	stack string
}

// Error implements the error interface.
// It ONLY returns the original error message, keeping logs clean.
func (e *detailedError) Error() string {
	return e.cause.Error()
}

// Unwrap allows errors.Is and errors.As to work perfectly.
func (e *detailedError) Unwrap() error {
	return e.cause
}

// StackTrace exposes the stack trace to loggers that know to look for it.
func (e *detailedError) StackTrace() string {
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

	return &detailedError{
		cause: cause,
		stack: runtimez.StackTraceString(3),
	}
}

// StackTrace searches for a DetailedError in the error chain and returns its
// stack trace. If no DetailedError is found, it returns an empty string.
func StackTrace(err error) string {
	if err == nil {
		return ""
	}

	var de DetailedError
	if errors.As(err, &de) {
		return de.StackTrace()
	}

	return ""
}
