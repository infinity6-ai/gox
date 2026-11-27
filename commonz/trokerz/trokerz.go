package trokerz

import (
	"context"
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

type Options struct {
	Ctx            context.Context
	Begin          string
	End            string
	Get            func(k string) (string, bool, error)
	IgnoreUnclosed bool
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

	ctx := opts.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	escapedBegin := opts.Begin + opts.Begin
	escapedEnd := opts.End + opts.End

	if opts.Begin == opts.End {
		if !strings.Contains(original, opts.Begin) {
			return original, nil
		}
	} else {
		if !strings.Contains(original, opts.Begin) && !strings.Contains(original, escapedEnd) {
			return original, nil
		}
	}

	var builder strings.Builder
	builder.Grow(len(original))
	remaining := original

	for {
		nextBegin := strings.Index(remaining, opts.Begin)

		if opts.Begin != opts.End {
			nextEnd := strings.Index(remaining, opts.End)
			if nextEnd != -1 && (nextBegin == -1 || nextEnd < nextBegin) {
				if strings.HasPrefix(remaining[nextEnd:], escapedEnd) {
					builder.WriteString(remaining[:nextEnd])
					builder.WriteString(opts.End)
					remaining = remaining[nextEnd+len(escapedEnd):]
					continue
				}
			}
		}

		if nextBegin == -1 {
			builder.WriteString(remaining)
			break
		}

		if strings.HasPrefix(remaining[nextBegin:], escapedBegin) {
			builder.WriteString(remaining[:nextBegin])
			builder.WriteString(opts.Begin)
			remaining = remaining[nextBegin+len(escapedBegin):]
			continue
		}

		builder.WriteString(remaining[:nextBegin])

		afterBegin := remaining[nextBegin+len(opts.Begin):]

		endIdx := strings.Index(afterBegin, opts.End)
		if endIdx == -1 {
			if opts.IgnoreUnclosed {
				builder.WriteString(opts.Begin)
				builder.WriteString(afterBegin)
				break
			}
			return "", fmt.Errorf("%w: missing %q", ErrUnclosedPlaceholder, opts.End)
		}

		if opts.Begin != opts.End {
			if strings.Contains(afterBegin[:endIdx], opts.Begin) {
				if opts.IgnoreUnclosed {
					builder.WriteString(opts.Begin)
					remaining = afterBegin
					continue
				}
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
