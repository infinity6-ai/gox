package logzctx

import (
	"context"
	"maps"
)

type kt int

const t kt = 1

func Get(ctx context.Context) map[string]string {
	ret := ctx.Value(t)
	if ret == nil {
		return nil
	}
	return maps.Clone(ret.(map[string]string))
}

func Set(ctx context.Context, attrs map[string]string) context.Context {
	if len(attrs) == 0 {
		return ctx
	}
	prev := Get(ctx)
	current := make(map[string]string)
	if prev != nil {
		maps.Copy(current, prev)
	}
	maps.Copy(current, attrs)
	ctx = context.WithValue(ctx, t, current)
	return ctx
}
