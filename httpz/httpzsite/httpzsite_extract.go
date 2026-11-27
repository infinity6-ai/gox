package httpzsite

import (
	"context"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/staticz"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzentry"
	"github.com/infinity6-ai/gox/commonz/trokerz"
)

// ExtractOptions specifies options for extracting a site's static assets.
type ExtractOptions struct {
	Path  *pathz.Path
	Troke bool
}

// ExtractTo extracts the site's static pack assets to the specified destination path.
func (s *Site) ExtractTo(ctx context.Context, opts ExtractOptions) error {
	if opts.Path == nil {
		return fmt.Errorf("failed to extract site pack: path is required")
	}

	if !opts.Troke || len(s.TrokerOptions) == 0 {
		err := staticz.ExtractTo(ctx, opts.Path, s.Pack)
		if err != nil {
			return fmt.Errorf("failed to extract site pack to %s: %w", opts.Path, err)
		}
		return nil
	}

	err := staticz.Walk(ctx, s.Pack, func(entry staticzentry.Entry) error {
		destPath := opts.Path.MustJoin(entry.Name())
		if err := os.MkdirAll(destPath.Dir().String(), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory for static entry %s at %s: %w", entry.Name(), destPath.Dir(), err)
		}
		r, err := entry.Open()
		if err != nil {
			return fmt.Errorf("failed to open staticz entry %s: %w", entry.Name(), err)
		}
		defer r.Close()

		data, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read staticz entry %s: %w", entry.Name(), err)
		}

		if utf8.Valid(data) {
			content := string(data)
			for _, opt := range s.TrokerOptions {
				content, err = trokerz.Troke(content, opt)
				if err != nil {
					return fmt.Errorf("failed to apply troker to %s: %w", entry.Name(), err)
				}
			}
			data = []byte(content)
		}

		w, err := os.OpenFile(destPath.String(), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
		}
		defer w.Close()

		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("failed to write destination file %s: %w", destPath, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to extract site pack to %s: %w", opts.Path, err)
	}
	return nil
}
