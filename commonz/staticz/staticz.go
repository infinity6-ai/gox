package staticz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzentry"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzlocal"
)

func Walk(ctx context.Context, name any, callback func(entry staticzentry.Entry) error) error {
	err := staticzlocal.Walk(ctx, name, callback)
	if !errors.Is(err, staticzlocal.ErrNotFound) {
		return err
	}
	return staticzloader.Walk(ctx, name, callback)
}

func Lookup(ctx context.Context, name any, p *pathz.Path) (staticzentry.Entry, error) {
	ret, err := staticzlocal.Lookup(ctx, name, p)
	if err != nil {
		return nil, fmt.Errorf("error looking up for local static file %s: %w", p, err)
	}
	if ret != nil {
		return ret, nil
	}
	return staticzloader.Lookup(ctx, name, p)
}

func ExtractTo(ctx context.Context, dest *pathz.Path, name ...any) error {
	for _, n := range name {
		err := Walk(ctx, n, func(entry staticzentry.Entry) error {
			destPath := dest.MustJoin(entry.Name())
			if err := os.MkdirAll(destPath.Dir().String(), os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory for static entry %s at %s: %w", entry.Name(), destPath.Dir(), err)
			}
			r, err := entry.Open()
			if err != nil {
				return fmt.Errorf("failed to open staticz entry %s: %w", entry.Name(), err)
			}
			defer r.Close()
			w, err := os.OpenFile(destPath.String(), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create or open destination file %s for staticz entry %s: %w", destPath, entry.Name(), err)
			}
			defer w.Close()
			if _, err := io.Copy(w, r); err != nil {
				return fmt.Errorf("failed to copy content from staticz entry %s to %s: %w", entry.Name(), destPath, err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to extract staticz for name %v: %w", n, err)
		}
	}
	return nil
}
