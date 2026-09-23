package trokerz

import (
	"fmt"
	"strings"
)

type Options struct {
	Begin string
	End   string
	Get   func(k string) (string, bool, error)
}

func Troke(original string, opts Options) (string, error) {
	if opts.Begin == "" || opts.End == "" || opts.Get == nil {
		panic("begin, end and get are required")
	}

	var builder strings.Builder
	remaining := original

	for {
		beginIdx := strings.Index(remaining, opts.Begin)
		if beginIdx == -1 {
			builder.WriteString(remaining)
			break
		}

		builder.WriteString(remaining[:beginIdx])

		afterBegin := remaining[beginIdx+len(opts.Begin):]

		endIdx := strings.Index(afterBegin, opts.End)
		if endIdx == -1 {
			builder.WriteString(opts.Begin)
			builder.WriteString(afterBegin)
			break
		}

		k := afterBegin[:endIdx]

		replacement, shouldReplace, err := opts.Get(k)
		if err != nil {
			return "", fmt.Errorf("failed to get replacement for key %q: %w", k, err)
		}

		if shouldReplace {
			builder.WriteString(replacement)
		} else {
			builder.WriteString(opts.Begin)
			builder.WriteString(k)
			builder.WriteString(opts.End)
		}

		remaining = afterBegin[endIdx+len(opts.End):]
	}

	return builder.String(), nil
}
