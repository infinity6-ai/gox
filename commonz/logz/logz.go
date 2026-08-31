package logz

import (
	"context"
	"fmt"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/logz/logzlast"
	"github.com/infinity6-ai/gox/commonz/logz/logzprovider"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
)

type Type bool

type Logger interface {
	Error(ctx context.Context, op string, params map[string]any, errs ...error)
	Info(ctx context.Context, op string, params map[string]any, errs ...error)
	Debug(ctx context.Context, op string, params map[string]any, errs ...error)
	Appender() string
}

type Collector interface {
	Collector() func() []*logzspec.Entry
}

type loggerImpl struct {
	appender string
	provider logzspec.ProviderLogger
	last     *logzlast.LastEntries
}

func (l *loggerImpl) Collector() func() []*logzspec.Entry {
	if l.last == nil {
		l.last = &logzlast.LastEntries{}
	}
	return func() []*logzspec.Entry {
		return l.last.Entries()
	}
}

func (l *loggerImpl) Appender() string {
	return l.appender
}

func (l *loggerImpl) send(ctx context.Context, entry *logzspec.Entry) {
	if l.last != nil {
		l.last.Add(entry)
	}
	l.provider(ctx, entry)
}

func (l *loggerImpl) Debug(ctx context.Context, op string, params map[string]any, errs ...error) {
	l.send(ctx, logzspec.NewEntry(1, l.appender, logzspec.DEBUG, op, params, errs...))
}

func (l *loggerImpl) Error(ctx context.Context, op string, params map[string]any, errs ...error) {
	l.send(ctx, logzspec.NewEntry(1, l.appender, logzspec.ERROR, op, params, errs...))
}

func (l *loggerImpl) Info(ctx context.Context, op string, params map[string]any, errs ...error) {
	l.send(ctx, logzspec.NewEntry(1, l.appender, logzspec.INFO, op, params, errs...))
}

func Create(logger any) Logger {
	t := reflect.TypeOf(logger)
	if t.Name() != "tlogger" {
		panic(fmt.Sprintf("appender type must be tlogger, but was: %s", t.Name()))
	}
	if fmt.Sprintf("%t", logger) != "true" {
		panic("logger must be true")
	}
	impl := &loggerImpl{}
	impl.appender = t.PkgPath()
	impl.provider = logzprovider.GetDefaultProvider()
	return impl
	// prv := logzprovider.GetDefaultProvider()
	// ret := prv(t.PkgPath())
	// ret = logzlast.New(ret)
	// return ret
}
