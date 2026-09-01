package timerz_test

import (
	"context"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/syncz/timerz"
	"github.com/stretchr/testify/assert"
)

func TestUnitDelayFor_Completes(t *testing.T) {
	ctx := context.Background()
	called := false
	fn := func() {
		called = true
	}

	start := time.Now()
	timerz.DelayFor(ctx, 20*time.Millisecond, fn)
	duration := time.Since(start)

	assert.True(t, called, "The function should have been called")
	assert.GreaterOrEqual(t, duration, 20*time.Millisecond, "The delay should be at least the specified duration")
}

func TestUnitDelayFor_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	called := false
	fn := func() {
		called = true
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	timerz.DelayFor(ctx, 50*time.Millisecond, fn)
	assert.False(t, called, "The function should not be called when context is canceled")
}

func TestUnitDelayUntil_Completes(t *testing.T) {
	ctx := context.Background()
	called := false
	fn := func() {
		called = true
	}

	until := time.Now().Add(20 * time.Millisecond)
	timerz.DelayUntil(ctx, until, fn)

	assert.True(t, called, "The function should have been called")
	assert.True(t, time.Now().After(until) || time.Now().Equal(until), "The current time should be after or at the until time")
}

func TestUnitDelayUntil_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	called := false
	fn := func() {
		called = true
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	timerz.DelayUntil(ctx, time.Now().Add(50*time.Millisecond), fn)
	assert.False(t, called, "The function should not be called when context is canceled")
}

func TestUnitDelayUntil_Past(t *testing.T) {
	ctx := context.Background()
	called := false
	timerz.DelayUntil(ctx, time.Now().Add(-10*time.Millisecond), func() { called = true })
	assert.True(t, called, "The function should have been called immediately for a past time")
}
