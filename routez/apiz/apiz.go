package apiz

import (
	"context"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/schemaz/schemazv2"
)

type DataRefs struct {
	PathParams  *schemazv2.Schema
	QueryParams *schemazv2.Schema
	ReqHeaders  *schemazv2.Schema
	ReqBody     *schemazv2.Schema
	RespHeaders *schemazv2.Schema
	RespBody    *schemazv2.Schema
}

type Api interface {
	GetDataRefs() *DataRefs
}

type Handler[T Api] func(ctx context.Context, reqResp T) (int, error)

type Service[T Api] struct {
	Spec    *Spec[T]
	Handler Handler[T]
}

type Spec[T Api] struct {
	Id     string
	Desc   func() *schemazv2.Desc
	Method string
	Path   string
	Spec   T
}

func (a *Spec[T]) MewReqResp() T {
	var v T
	t := reflect.TypeOf(&v).Elem()
	checker.Equal(reflect.Ptr, t.Kind(), "it must be a pointer: %T %T", v, t)
	ret := reflect.New(t.Elem()).Interface().(T)
	return ret
}
