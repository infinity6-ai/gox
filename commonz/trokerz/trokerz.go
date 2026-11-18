package trokerz

import (
	"errors"
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

// ErrInvalidOptions is returned when Begin, End, or Get is missing in Options.
var ErrInvalidOptions = errors.New("begin, end and get are required")

type Options struct {
	Begin string
	End   string
	Get   func(k string) (string, bool, error)
}

func MustTroke(original string, opts Options) string {
	ret, err := Troke(original, opts)
	errorz.Check(err)
	return ret
}

func Troke(original string, opts Options) (string, error) {
	if opts.Begin == "" || opts.End == "" || opts.Get == nil {
		return "", ErrInvalidOptions
	}

	if !strings.Contains(original, opts.Begin) {
		return original, nil
	}

	var builder strings.Builder
	builder.Grow(len(original))
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

		if opts.Begin != opts.End {
			if lastBegin := strings.LastIndex(afterBegin[:endIdx], opts.Begin); lastBegin != -1 {
				builder.WriteString(opts.Begin)
				builder.WriteString(afterBegin[:lastBegin])
				afterBegin = afterBegin[lastBegin+len(opts.Begin):]
				endIdx -= lastBegin + len(opts.Begin)
			}
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
