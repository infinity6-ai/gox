package errorz

import (
	"fmt"
	"runtime"
	"strings"
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

// New wraps the error and captures the stack.
func New(cause error) error {
	if cause == nil {
		return nil
	}

	// getStackTrace() would be the same helper function from the previous response
	return &customError{
		cause: cause,
		stack: getStackTrace(),
	}
}

// getStackTrace captures the call stack of the current goroutine.
func getStackTrace() string {
	const maxDepth = 32
	var pcs [maxDepth]uintptr

	// Skip 3 frames to avoid including runtime.Callers, getStackTrace, and errorz.New
	// in the final output.
	n := runtime.Callers(3, pcs[:])
	if n == 0 {
		return "no stack trace available"
	}

	frames := runtime.CallersFrames(pcs[:n])
	var sb strings.Builder

	for {
		frame, more := frames.Next()
		// Format: /path/to/file.go:line_number (Function.Name)
		fmt.Fprintf(&sb, "\t%s:%d (%s)\n", frame.File, frame.Line, frame.Function)

		if !more {
			break
		}
	}

	return strings.TrimSuffix(sb.String(), "\n")
}
