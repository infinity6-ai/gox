package apiz

import (
	"context"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type DataRefs struct {
	PathParams  any
	QueryParams any
	ReqHeaders  any
	ReqBody     any
	RespHeaders any
	RespBody    any
}

type ReqResp interface {
	GetDataRefs() *DataRefs
}

type Handler[T ReqResp] func(ctx context.Context, reqResp T) (int, error)

type Api[T ReqResp] struct {
	Schema  *schemaz.Api
	Handler Handler[T]
}

func (a *Api[T]) MewReqResp() T {
	var v T
	t := reflect.TypeOf(&v).Elem()
	checker.Equal(reflect.Ptr, t.Kind(), "it must be a pointer: %T %T", v, t)
	ret := reflect.New(t.Elem()).Interface().(T)
	return ret
}
