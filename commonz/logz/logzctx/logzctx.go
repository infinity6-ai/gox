package logzctx

import (
	"context"
	"maps"
)

type kt int

const t kt = 1

func gatMap(ctx context.Context) map[string]string {
	ret := ctx.Value(t)
	if ret == nil {
		return nil
	}
	return ret.(map[string]string)
}

func Set(ctx context.Context, attrs map[string]string) context.Context {
	m := gatMap(ctx)
	if m == nil {
		m = make(map[string]string)
		ctx = context.WithValue(ctx, t, m)
	}
	maps.Copy(m, attrs)
	return ctx
}

func Get(ctx context.Context) map[string]string {
	return maps.Clone(gatMap(ctx))
}
