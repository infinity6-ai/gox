package apiz

import (
	"context"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/schemaz/schemazv2"
)

type DataRefsV2 struct {
	PathParams  *schemazv2.Schema
	QueryParams *schemazv2.Schema
	ReqHeaders  *schemazv2.Schema
	ReqBody     *schemazv2.Schema
	RespHeaders *schemazv2.Schema
	RespBody    *schemazv2.Schema
}

type ReqRespV2 interface {
	GetDataRefsV2() *DataRefsV2
}

type HandlerV2[T ReqRespV2] func(ctx context.Context, reqResp T) (int, error)

type ApiV2[T ReqRespV2] struct {
	Id      string
	Desc    func() *schemazv2.Desc
	Method  string
	Path    string
	Spec    T
	Handler HandlerV2[T]
}

func (a *ApiV2[T]) MewReqResp() T {
	var v T
	t := reflect.TypeOf(&v).Elem()
	checker.Equal(reflect.Ptr, t.Kind(), "it must be a pointer: %T %T", v, t)
	ret := reflect.New(t.Elem()).Interface().(T)
	return ret
}
