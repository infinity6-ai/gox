package errorz

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/runtimez"
)

// Data is the structured, machine-readable data of a StructuredError.
type Data struct {
	Code    int    `json:"code,omitempty"`
	Name    string `json:"name,omitempty"`
	Payload string `json:"payload,omitempty"`
	Cause   string `json:"cause"`
}

// structuredError holds the original error and the structured data.
type structuredError struct {
	cause   error
	stack   string
	code    int
	name    string
	payload string
}

// toCoolString returns a detailed string with all structured fields.
// This is used as a fallback for Error() if JSON marshaling fails.
func (e *structuredError) toCoolString() string {
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

	if len(details) > 0 {
		sb.WriteString(": (")
		sb.WriteString(strings.Join(details, ", "))
		sb.WriteString(")")
	}

	return sb.String()
}

// Error implements the error interface.
// It returns a JSON string representation of the error's Data.
func (e *structuredError) Error() string {
	jsonData, err := json.Marshal(e.Data())
	if err != nil {
		// Fallback to a detailed string format if JSON marshaling fails.
		return e.toCoolString()
	}
	return string(jsonData)
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
		Code:    e.code,
		Name:    e.name,
		Payload: e.payload,
		Cause:   e.cause.Error(),
	}
}

// StructuredError is an error with a stack trace and structured data.
type StructuredError interface {
	error
	Unwrap() error
	StackTrace() string
	Code() int
	Name() string
	Payload() string
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
