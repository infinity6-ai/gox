package errorz

import (
	"errors"
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/runtimez"
)

// structuredError holds the original error and the structured data.
type structuredError struct {
	cause    error
	stack    string
	code     int
	name     string
	payload  string
	business bool
}

// Error implements the error interface.
// It returns a detailed, human-readable string with all structured fields.
func (e *structuredError) Error() string {
	var sb strings.Builder
	if e.cause != nil {
		sb.WriteString(e.cause.Error())
	}

	var details []string
	if e.name != "" {
		details = append(details, e.name)
	}
	if e.code != 0 {
		details = append(details, fmt.Sprintf("code=%d", e.code))
	}
	if e.payload != "" {
		details = append(details, fmt.Sprintf("payload=%s", e.payload))
	}
	if e.business {
		details = append(details, "business=true")
	}

	if len(details) > 0 {
		sb.WriteString(": (")
		sb.WriteString(strings.Join(details, ", "))
		sb.WriteString(")")
	}

	return sb.String()
}

// Unwrap allows errors.Is and errors.As to work perfectly.
func (e *structuredError) Unwrap() error {
	return e.cause
}

// StackTrace exposes the stack trace to loggers that know to look for it.
func (e *structuredError) StackTrace() string {
	return e.stack
}

// Data returns the structured, machine-readable data of the error.
func (e *structuredError) Data() *Data {
	if e == nil {
		return nil
	}
	return &Data{
		Code:     e.code,
		Name:     e.name,
		Payload:  e.payload,
		Business: e.business,
		Cause:    e.cause.Error(),
		Stack:    e.stack,
	}
}

// Business returns true if the error is intended for the end-user.
func (e *structuredError) Business() bool {
	return e.business
}

// StructuredError is an error with a stack trace and structured data.
type StructuredError interface {
	error
	Unwrap() error
	StackTrace() string
	Code() int
	Name() string
	Payload() string
	Business() bool
	Data() *Data
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
func Detail(code int, name string, payload string, business bool, cause error) StructuredError {
	if cause == nil {
		return nil
	}

	return &structuredError{
		cause:    cause,
		stack:    runtimez.StackTraceString(3),
		code:     code,
		name:     name,
		payload:  payload,
		business: business,
	}
}

// Detailf creates a new error from a format string and arguments, then wraps it
// with a stack trace and structured data.
func Detailf(code int, name string, payload string, business bool, format string, a ...interface{}) StructuredError {
	cause := fmt.Errorf(format, a...)
	return &structuredError{
		cause:    cause,
		stack:    runtimez.StackTraceString(3),
		code:     code,
		name:     name,
		payload:  payload,
		business: business,
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
