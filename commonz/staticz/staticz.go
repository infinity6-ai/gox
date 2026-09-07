package staticz

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzentry"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzlocal"
	"go.code.infinity6.ai/platform/util/ioz"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

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
	code, err := staticzloader.Lookup(ctx, name, p)
	if err != nil {
		return nil, fmt.Errorf("error looking up for code static file %s: %w", p, err)
	}
	if ret == nil {
		ret = code
	}
	if ret == nil {
		return nil, nil
	}
	err = compare(ctx, ret, code)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func compare(ctx context.Context, ret, code staticzentry.Entry) error {
	localReader, err := ret.Open()
	if err != nil {
		return err
	}
	localContent := ioz.MustReadLimited(localReader, max)
	if code == nil {
		logger.Info(ctx, "static file found locally only", map[string]any{"f": ret.Name()})
		return nil
	}
	codeReader, err := code.Open()
	if err != nil {
		return err
	}
	codeContent := ioz.MustReadLimited(codeReader, max)
	if !bytes.Equal(localContent, codeContent) {
		logger.Info(ctx, "static file does not match", map[string]any{"f": ret.Name()})
	}
	return nil
}

type Pack interface {
	Name() any
	String() string
}
