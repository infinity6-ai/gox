// Package runtimez provides utility functions for interacting with the Go runtime,
// such as capturing and formatting stack traces.
package runtimez

import (
	"fmt"
	"runtime"
	"strings"
)

// StackTraceString captures the call stack of the current goroutine and formats it
// as a multi-line string.
//
// The 'skip' parameter indicates the number of stack frames to ascend for
// runtime.Callers, with 0 identifying the frame for runtime.Callers itself.
// To start the stack trace from the caller of the function that calls
// StackTraceString, you need to provide a skip value that accounts for the
// call stack depth. For instance, if `foo()` calls `bar()`, and `bar()` calls
// `StackTraceString()`, to have the trace start at `foo()`, you would need a
// skip value of 3 (to skip runtime.Callers, StackTraceString, and bar).
//
// The output is formatted as a list of lines, where each line has the format:
//
//	/path/to/file.go:line_number (package.FunctionName)
func StackTraceString(skip int) string {
	const maxDepth = 32
	var pcs [maxDepth]uintptr

	n := runtime.Callers(skip, pcs[:])
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
