package apiz

import (
	"context"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type DataRefs struct {
	PathParams  *schemaz.Schema
	QueryParams *schemaz.Schema
	ReqHeaders  *schemaz.Schema
	ReqBody     *schemaz.Schema
	RespHeaders *schemaz.Schema
	RespBody    *schemaz.Schema
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
	Desc   func() *schemaz.Desc
	Method string
	Path   string
	Spec   T
}

func (a *Spec[T]) NewReqResp() T {
	var v T
	t := reflect.TypeOf(&v).Elem()
	checker.Equal(reflect.Pointer, t.Kind(), "it must be a pointer: %T %T", v, t)
	ret := reflect.New(t.Elem()).Interface().(T)
	return ret
}
