package runtimez

import (
	"fmt"
	"runtime"
	"strings"
)

// StackTraceString captures the call stack of the current goroutine.
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
