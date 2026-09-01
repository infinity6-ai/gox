package promise

import (
	"context"
	"fmt"
	"sync"

	"github.com/infinity6-ai/gox/commonz/constraintz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/logz"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

type PromisePanicError struct {
	Panic any
}

func (me *PromisePanicError) Error() string {
	err, ok := me.Panic.(error)
	if !ok {
		return fmt.Sprintf("promise has panic'ed: %#v", me.Panic)
	}
	return fmt.Sprintf("promise has panic'ed with error: %s, %#v", err.Error(), me.Panic)
}

type PromiseWasNotResolvedError struct {
}

func (me *PromiseWasNotResolvedError) Error() string {
	return "promise was not resolved"
}

type Promise[T any] struct {
	ctx   context.Context
	once  sync.Once
	value T
	err   error
	done  chan struct{}
}

func New[T any](ctx context.Context) *Promise[T] {
	return &Promise[T]{ctx: ctx, done: make(chan struct{})}
}

func (me *Promise[T]) ResolveDeferrable() {
	p := recover()
	var err error
	if p == nil {
		if me.IsResolved() {
			return
		}
		err = &PromiseWasNotResolvedError{}
	} else {
		err = &PromisePanicError{Panic: p}
		logger.Error(me.ctx, "error captured on async func", nil, err)
	}
	var value T
	me.Resolve(value, err)
}

func (me *Promise[T]) Resolve(value T, err error) {
	me.once.Do(func() {
		me.value = value
		me.err = err
		close(me.done)
	})
}

func (me *Promise[T]) Get() (T, error) {
	<-me.done
	return me.value, me.err
}

func (me *Promise[T]) GetV() T {
	ret, err := me.Get()
	errorz.Check(err)
	return ret
}

func (me *Promise[T]) WaitDeferrable() {
	_, err := me.Get()
	if err != nil {
		logger.Error(me.ctx, "close wait", nil, err)
	}
}

func (me *Promise[T]) IsResolved() bool {
	select {
	case <-me.done:
		return true
	default:
		return false
	}
}

func Async[T any](ctx context.Context, fn func() (T, error)) *Promise[T] {
	p := New[T](ctx)
	go func() {
		defer p.ResolveDeferrable()
		val, err := fn()
		p.Resolve(val, err)
	}()
	return p
}

func AsyncV(ctx context.Context, fn func()) *Promise[constraintz.Void] {
	return Async(ctx, func() (constraintz.Void, error) {
		fn()
		return nil, nil
	})
}

func Wait[T any](promises ...*Promise[T]) {
	for _, p := range promises {
		p.Get()
	}
}
