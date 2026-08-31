package logz

import (
	"context"
	"fmt"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/logz/logzevent"
	"github.com/infinity6-ai/gox/commonz/logz/logzprovider"
)

type Type bool

type Logger interface {
	Error(ctx context.Context, op string, params map[string]any, errs ...error)
	Info(ctx context.Context, op string, params map[string]any, errs ...error)
	Debug(ctx context.Context, op string, params map[string]any, errs ...error)
	Appender() string
}

func Create(logger any) Logger {
	t := reflect.TypeOf(logger)
	if t.Name() != "tlogger" {
		panic(fmt.Sprintf("appender type must be tlogger, but was: %s", t.Name()))
	}
	if fmt.Sprintf("%t", logger) != "true" {
		panic("logger must be true")
	}
	prv := logzprovider.GetDefaultProvider()
	ret := prv(t.PkgPath())
	ret = logzevent.New(ret)
	return ret
}
