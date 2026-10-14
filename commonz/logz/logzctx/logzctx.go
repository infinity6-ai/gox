package logzctx

import (
	"context"
	"maps"
)

type kt int

const t kt = 1

func Set(ctx context.Context, attrs map[string]string) context.Context {
	m := ctx.Value(t).(map[string]string)
	if m == nil {
		m = make(map[string]string)
		ctx = context.WithValue(ctx, t, m)
	}
	maps.Copy(m, attrs)
	return ctx
}

func Get(ctx context.Context, k string) map[string]string {
	m := ctx.Value(t).(map[string]string)
	return maps.Clone(m)
}
