package main

import (
	"strings"
)

type Options struct {
	Begin string
	End   string
	Get   func(v string) (string, bool, error)
}

func Troke(original string, opts Options) (string, error) {
	// Fast exit if delimiters are empty
	if opts.Begin == "" || opts.End == "" {
		return original, nil
	}

	var builder strings.Builder
	remaining := original

	for {
		// Find the next opening delimiter
		beginIdx := strings.Index(remaining, opts.Begin)
		if beginIdx == -1 {
			// No more begin delimiters, append the rest and finish
			builder.WriteString(remaining)
			break
		}

		// Append the text before the delimiter
		builder.WriteString(remaining[:beginIdx])

		// Focus on the string after the Begin delimiter
		afterBegin := remaining[beginIdx+len(opts.Begin):]

		// Find the matching closing delimiter
		endIdx := strings.Index(afterBegin, opts.End)
		if endIdx == -1 {
			// No matching end delimiter, append the unmatched part and finish
			builder.WriteString(opts.Begin)
			builder.WriteString(afterBegin)
			break
		}

		// Extract the value inside the delimiters
		v := afterBegin[:endIdx]

		// Retrieve the replacement value, the "should replace" flag, and potential error
		replacement, shouldReplace, err := opts.Get(v)
		if err != nil {
			// Fail fast and return the error to the caller
			return "", err
		}

		if shouldReplace {
			// Apply the replacement
			builder.WriteString(replacement)
		} else {
			// "Don't touch this": Restore the original placeholder verbatim
			builder.WriteString(opts.Begin)
			builder.WriteString(v)
			builder.WriteString(opts.End)
		}

		// Advance the remaining string past the End delimiter
		remaining = afterBegin[endIdx+len(opts.End):]
	}

	return builder.String(), nil
}
