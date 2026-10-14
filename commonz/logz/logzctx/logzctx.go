package logzctx

import (
	"context"
)

type kt int

const t kt = 1

func Prepare(ctx context.Context) context.Context {
	m := ctx.Value(t)
	if m == nil {
		ctx = context.WithValue(ctx, t, make(map[string]string))
	}
	return ctx
}

func Get(ctx context.Context) map[string]string {
	ret := ctx.Value(t)
	if ret == nil {
		return nil
	}
	return ret.(map[string]string)
}
