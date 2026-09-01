package timerz

import (
	"context"
	"time"
)

func DelayFor(ctx context.Context, duration time.Duration, fn func()) {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-timer.C:
		fn()
	case <-ctx.Done():
	}
}

func DelayUntil(ctx context.Context, until time.Time, fn func()) {
	duration := time.Until(until)
	DelayFor(ctx, duration, fn)
}
