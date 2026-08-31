package logz

import (
	"fmt"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/logz/logzlast"
	"github.com/infinity6-ai/gox/commonz/logz/logzprovider"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
)

type Type bool

func Create(logger any) logzspec.Logger {
	t := reflect.TypeOf(logger)
	if t.Name() != "tlogger" {
		panic(fmt.Sprintf("appender type must be tlogger, but was: %s", t.Name()))
	}
	if fmt.Sprintf("%t", logger) != "true" {
		panic("logger must be true")
	}
	prv := logzprovider.GetDefaultProvider()
	ret := prv(t.PkgPath())
	ret = logzlast.New(ret)
	return ret
}
