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

func ExtractTo(ctx context.Context, name any, dest *pathz.Path) error {
	return Walk(ctx, name, func(entry staticzentry.Entry) error {
		destPath := dest.MustJoin(entry.Name())
		if err := os.MkdirAll(destPath.Dir().String(), os.ModePerm); err != nil {
			return err
		}
		r, err := entry.Open()
		if err != nil {
			return err
		}
		defer r.Close()
		w, err := os.OpenFile(destPath.String(), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = io.Copy(w, r)
		return err
	})
}
