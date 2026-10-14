package logzctx

import (
	"context"
	"maps"
)

type kt int

const t kt = 1

func Prepare(ctx context.Context) context.Context {
	m := Get(ctx)
	newMap := make(map[string]string, len(m))
	if m != nil {
		maps.Copy(newMap, m)
	}
	return context.WithValue(ctx, t, newMap)
}

func Get(ctx context.Context) map[string]string {
	ret := ctx.Value(t)
	if ret == nil {
		return nil
	}
	return ret.(map[string]string)
}
