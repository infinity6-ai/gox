package errorz

import (
	"errors"
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/runtimez"
)

// Data is the structured, machine-readable data of a StructuredError.
type Data struct {
	Code     int    `json:"code,omitempty"`
	Name     string `json:"name,omitempty"`
	Payload  string `json:"payload,omitempty"`
	Business bool   `json:"business,omitempty"`
	Cause    string `json:"cause"`
	Stack    string `json:"stack,omitempty"`
}

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

// BusinessData sanitizes an error for business-level display and returns its Data struct.
// If the error is not a StructuredError, it wraps it as a generic internal error
// and returns its Data.
// If it's a StructuredError marked as business-facing, it returns its Data as is.
// If it's a StructuredError but NOT business-facing, it hides internal details
// by creating a new StructuredError with a generic "InternalError" message,
// retaining the original code and name if present, and then returns its Data.
func BusinessData(err error) *Data {
	if err == nil {
		return nil
	}

	var se StructuredError
	var initialData *Data // This will hold the initial Data before stripping

	if !errors.As(err, &se) {
		// Not a StructuredError, wrap as generic internal error, then get its data.
		sanitizedErr := Detail(500, "InternalError", "", false, err)
		initialData = sanitizedErr.Data()
	} else if se.Business() {
		// Already a business-facing error, return its data.
		initialData = se.Data()
	} else {
		// StructuredError but not business-facing, hide internal details, then get its data.
		// We retain Code and Name if available from the original StructuredError,
		// but replace the cause with a generic message and mark as non-business.
		sanitizedErr := Detail(se.Code(), se.Name(), "", false, errors.New("InternalError"))
		initialData = sanitizedErr.Data()
	}

	// Apply the stripping logic
	initialData.Stack = "" // Always remove stack

	if initialData.Business { // If it's a business error after sanitization
		initialData.Payload = "" // Strip payload for business errors
	}

	return initialData
}
