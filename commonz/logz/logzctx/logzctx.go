package logzctx

import (
	"context"
	"maps"
)

type kt int

const t kt = 1

func Set(ctx context.Context, k string, v string) context.Context {
	m := ctx.Value(t).(map[string]string)
	if m == nil {
		m = make(map[string]string)
		ctx = context.WithValue(ctx, t, m)
	}
	m[k] = v
	return ctx
}

func All(ctx context.Context, k string) map[string]string {
	m := ctx.Value(t).(map[string]string)
	return maps.Clone(m)
}
