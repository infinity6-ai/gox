package trokerz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/logz"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

// Logger returns the package-level logger.
func Logger() logz.Logger {
	return logger
}

// ErrInvalidOptions is returned when Begin, End, or Get is missing in Options.
var ErrInvalidOptions = errors.New("begin, end and get are required")

// ErrUnclosedPlaceholder is returned when an opening delimiter has no matching closing delimiter,
// or when opening delimiters are interleaved without being closed.
var ErrUnclosedPlaceholder = errors.New("unclosed placeholder")

// ErrUnmatchedClosingDelimiter is returned when a closing delimiter is found without a preceding opening delimiter.
var ErrUnmatchedClosingDelimiter = errors.New("unmatched closing delimiter")

type Options struct {
	Ctx   context.Context
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

	ctx := opts.Ctx
	if ctx == nil {
		ctx = context.Background()
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

	escapedBegin := opts.Begin + opts.Begin
	escapedEnd := opts.End + opts.End

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
				return "", fmt.Errorf("%w: found %q without preceding %q", ErrUnmatchedClosingDelimiter, opts.End, opts.Begin)
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
			logger.Info(ctx, "replace placeholder", map[string]any{
				"key":         k,
				"replacement": replacement,
			})
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
