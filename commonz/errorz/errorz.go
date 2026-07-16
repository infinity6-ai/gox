package errorz

import (
	"errors"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/runtimez"
)

// structuredError holds the original error and the structured data.
type structuredError struct {
	cause   error
	stack   string
	code    int
	name    string
	payload string
}

// Error implements the error interface.
// It ONLY returns the original error message, keeping logs clean.
func (e *structuredError) Error() string {
	return e.cause.Error()
}

// Unwrap allows errors.Is and errors.As to work perfectly.
func (e *structuredError) Unwrap() error {
	return e.cause
}

// StackTrace exposes the stack trace to loggers that know to look for it.
func (e *structuredError) StackTrace() string {
	return e.stack
}

// StructuredError is an error with a stack trace and structured data.
type StructuredError interface {
	error
	Unwrap() error
	StackTrace() string
	Code() int
	Name() string
	Payload() string
}

func (e *structuredError) Code() int {
	return e.code
}

func (e *structuredError) Name() string {
	return e.name
}

func (e *structuredError) Payload() string {
	return e.payload
}

// Detail wraps the error, captures the stack, and adds structured data.
// If the cause is nil, Detail returns nil.
func Detail(code int, name string, payload string, cause error) StructuredError {
	if cause == nil {
		return nil
	}

	return &structuredError{
		cause:   cause,
		stack:   runtimez.StackTraceString(3),
		code:    code,
		name:    name,
		payload: payload,
	}
}

// Detailf creates a new error from a format string and arguments, then wraps it
// with a stack trace and structured data.
func Detailf(code int, name string, payload string, format string, a ...interface{}) StructuredError {
	cause := fmt.Errorf(format, a...)
	return &structuredError{
		cause:   cause,
		stack:   runtimez.StackTraceString(3),
		code:    code,
		name:    name,
		payload: payload,
	}
}


// StackTrace searches for a StructuredError in the error chain and returns its
// stack trace. If no StructuredError is found, it returns an empty string.
func StackTrace(err error) string {
	if err == nil {
		return ""
	}

	var se StructuredError
	if errors.As(err, &se) {
		return se.StackTrace()
	}

	return ""
}
