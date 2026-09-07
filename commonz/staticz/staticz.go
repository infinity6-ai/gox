package staticz

import (
	"context"
	"errors"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzentry"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzlocal"
)

const max = 1 * 1024 * 1024

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
