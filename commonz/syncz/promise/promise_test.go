package promise_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/syncz/promise"
	"github.com/stretchr/testify/assert"
)

func TestUnitPromiseResolveAndGet(t *testing.T) {
	ctx := context.Background()
	p := promise.New[int](ctx)
	defer p.WaitDeferrable()
	assert.False(t, p.IsResolved())
	go func() {
		defer p.ResolveDeferrable()
		time.Sleep(100 * time.Millisecond)
		assert.False(t, p.IsResolved())
		p.Resolve(42, nil)
	}()

	assert.False(t, p.IsResolved())

	val, err := p.Get()
	assert.Nil(t, err)
	assert.Equal(t, 42, val)

	assert.True(t, p.IsResolved())

	val, err = p.Get()
	assert.Nil(t, err)
	assert.Equal(t, 42, val)

	assert.True(t, p.IsResolved())

}

func TestUnitPromiseResolveAndGetError(t *testing.T) {
	ctx := context.Background()
	p := promise.New[int](ctx)
	defer p.WaitDeferrable()
	go func() {
		defer p.ResolveDeferrable()
		p.Resolve(0, fmt.Errorf("myerror"))
	}()
	val, err := p.Get()
	assert.Equal(t, "myerror", err.Error())
	assert.Equal(t, 0, val)

	val, err = p.Get()
	assert.Equal(t, "myerror", err.Error())
	assert.Equal(t, 0, val)
}

func TestUnitPromiseResolveAndGetPanic(t *testing.T) {
	ctx := context.Background()
	p := promise.New[int](ctx)
	defer p.WaitDeferrable()
	assert.False(t, p.IsResolved())
	go func() {
		defer p.ResolveDeferrable()
		panic("othererror")
	}()

	assert.PanicsWithError(t, "promise has panic'ed: \"othererror\": (InternalError, code=500)", func() {
		p.GetV()
	})
	assert.True(t, p.IsResolved())
	assert.PanicsWithError(t, "promise has panic'ed: \"othererror\": (InternalError, code=500)", func() {
		p.GetV()
	})
	assert.True(t, p.IsResolved())
}

func TestUnitPromiseAsyncSuccess(t *testing.T) {
	ctx := context.Background()
	p := promise.Async(ctx, func() (string, error) {
		time.Sleep(50 * time.Millisecond)
		return "hello", nil
	})
	defer p.WaitDeferrable()
	val, err := p.Get()
	assert.Nil(t, err)
	assert.Equal(t, "hello", val)

	val, err = p.Get()
	assert.Nil(t, err)
	assert.Equal(t, "hello", val)
}

func TestUnitPromiseAsyncError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("boom")
	p := promise.Async(ctx, func() (int, error) {
		return 0, expectedErr
	})

	defer p.WaitDeferrable()

	val, err := p.Get()
	assert.Equal(t, "boom", err.Error())
	assert.Equal(t, 0, val)

	val, err = p.Get()
	assert.Equal(t, "boom", err.Error())
	assert.Equal(t, 0, val)
}

func TestUnitPromiseAsyncPanic(t *testing.T) {
	ctx := context.Background()
	p := promise.Async(ctx, func() (int, error) {
		panic("mypanic")
	})
	defer p.WaitDeferrable()
	assert.PanicsWithError(t, "promise has panic'ed: \"mypanic\": (InternalError, code=500)", func() {
		p.GetV()
	})

	assert.PanicsWithError(t, "promise has panic'ed: \"mypanic\": (InternalError, code=500)", func() {
		p.GetV()
	})
}

func TestUnitPromiseWait(t *testing.T) {
	ctx := context.Background()
	p1 := promise.Async(ctx, func() (string, error) {
		time.Sleep(50 * time.Millisecond)
		return "hello1", nil
	})
	defer p1.WaitDeferrable()
	p2 := promise.Async(ctx, func() (string, error) {
		time.Sleep(50 * time.Millisecond)
		return "hello2", nil
	})
	defer p2.WaitDeferrable()
	promise.Wait(p1, p2)

	assert.True(t, p1.IsResolved())
	assert.Equal(t, "hello1", p1.GetV())

	assert.True(t, p2.IsResolved())
	assert.Equal(t, "hello2", p2.GetV())
}

func TestUnitPromiseAsyncV(t *testing.T) {
	ctx := context.Background()
	x := "a1"
	p1 := promise.AsyncV(ctx, func() {
		time.Sleep(50 * time.Millisecond)
		x = "a2"
	})
	defer p1.WaitDeferrable()

	assert.False(t, p1.IsResolved())
	assert.Equal(t, "a1", x)

	assert.Nil(t, p1.GetV())
	assert.True(t, p1.IsResolved())
	assert.Equal(t, "a2", x)

}
