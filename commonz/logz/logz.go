package logz

import (
	"context"
	"fmt"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
)

type Type bool

type Logger interface {
	Error(ctx context.Context, op string, params map[string]any, errs ...error)
	Info(ctx context.Context, op string, params map[string]any, errs ...error)
	Debug(ctx context.Context, op string, params map[string]any, errs ...error)
	Appender() string
}

type loggerImpl struct {
	appender string
	provider logzspec.ProviderLogger
}

func (l *loggerImpl) Appender() string {
	return l.appender
}

func (l *loggerImpl) Debug(ctx context.Context, op string, params map[string]any, errs ...error) {
	l.provider(ctx, logzspec.NewEntry(1, l.appender, logzspec.DEBUG, op, params, errs...))
}

func (l *loggerImpl) Error(ctx context.Context, op string, params map[string]any, errs ...error) {
	l.provider(ctx, logzspec.NewEntry(1, l.appender, logzspec.ERROR, op, params, errs...))
}

func (l *loggerImpl) Info(ctx context.Context, op string, params map[string]any, errs ...error) {
	l.provider(ctx, logzspec.NewEntry(1, l.appender, logzspec.INFO, op, params, errs...))
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
	return impl
	// prv := logzprovider.GetDefaultProvider()
	// ret := prv(t.PkgPath())
	// ret = logzlast.New(ret)
	// return ret
}
