package trokerz

import (
	"errors"
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

// ErrInvalidOptions is returned when Begin, End, or Get is missing in Options.
var ErrInvalidOptions = errors.New("begin, end and get are required")

// ErrUnclosedPlaceholder is returned when an opening delimiter has no matching closing delimiter,
// or when opening delimiters are interleaved without being closed.
var ErrUnclosedPlaceholder = errors.New("unclosed placeholder")

// ErrUnmatchedClosingDelimiter is returned when a closing delimiter is found without a preceding opening delimiter.
var ErrUnmatchedClosingDelimiter = errors.New("unmatched closing delimiter")

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

	if opts.Begin == opts.End {
		if !strings.Contains(original, opts.Begin) {
			return original, nil
		}
	} else {
		if !strings.Contains(original, opts.Begin) && !strings.Contains(original, opts.End) {
			return original, nil
		}
	}

	var builder strings.Builder
	builder.Grow(len(original))
	remaining := original

	for {
		if opts.Begin != opts.End {
			endIdx := strings.Index(remaining, opts.End)
			beginIdx := strings.Index(remaining, opts.Begin)
			if endIdx != -1 && (beginIdx == -1 || endIdx < beginIdx) {
				return "", fmt.Errorf("%w: found %q without preceding %q", ErrUnmatchedClosingDelimiter, opts.End, opts.Begin)
			}
		}

		beginIdx := strings.Index(remaining, opts.Begin)
		if beginIdx == -1 {
			builder.WriteString(remaining)
			break
		}

		builder.WriteString(remaining[:beginIdx])

		afterBegin := remaining[beginIdx+len(opts.Begin):]

		endIdx := strings.Index(afterBegin, opts.End)
		if endIdx == -1 {
			return "", fmt.Errorf("%w: missing %q", ErrUnclosedPlaceholder, opts.End)
		}

		if opts.Begin != opts.End {
			if strings.Contains(afterBegin[:endIdx], opts.Begin) {
				return "", fmt.Errorf("%w: unclosed %q before %q", ErrUnclosedPlaceholder, opts.Begin, opts.End)
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
