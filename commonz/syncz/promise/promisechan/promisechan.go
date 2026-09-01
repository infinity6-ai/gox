package promisechan

import (
	"context"

	"github.com/infinity6-ai/gox/commonz/deferz"
	"github.com/infinity6-ai/gox/commonz/syncz/promise"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
)

func Async[I any, O any](ctx context.Context, in chan I, outBuffer int, numwokers int, fn func(out chan O)) (chan O, func()) {
	checker.Greater(numwokers, 0, "numwokers")
	checker.GreaterOrEqual(outBuffer, 0, "outBuffer")
	dfz := deferz.New(ctx)
	out := make(chan O, outBuffer)
	dfz.Add(func() {
		close(out)
	})
	for i := 0; i < numwokers; i++ {
		p := promise.AsyncV(ctx, func() {
			fn(out)
		})
		dfz.Add(p.WaitDeferrable)
	}
	end := promise.AsyncV(ctx, dfz.Do)
	return out, end.WaitDeferrable
}
