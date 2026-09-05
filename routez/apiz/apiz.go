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
	NewDataRefs()
	GetDataRefs() *DataRefs
}

type HandlerV2[T ReqResp] func(ctx context.Context, reqResp T) (int, error)

type ApiV2[T ReqResp] struct {
	Schema    *schemaz.Api
	HandlerV2 HandlerV2[T]
}

func (a *ApiV2[T]) MewReqResp() T {
	var v T
	t := reflect.TypeOf(&v).Elem()
	checker.Equal(reflect.Ptr, t.Kind(), "it must be a pointer: %T %T", v, t)
	ret := reflect.New(t.Elem()).Interface().(T)
	ret.NewDataRefs()
	return ret
}

type Req[P any, Q any, IH any, IB any] struct {
	PathParams  P
	QueryParams Q
	ReqHeaders  IH
	ReqBody     IB
}

type Resp[OH any, OB any] struct {
	Status      int
	RespHeaders OH
	RespBody    OB
}

type Handler[P any, Q any, IH any, IB any, OH any, OB any] func(ctx context.Context, req *Req[P, Q, IH, IB]) (*Resp[OH, OB], error)

type Api[P any, Q any, IH any, IB any, OH any, OB any] struct {
	Schema  *schemaz.Api
	Handler Handler[P, Q, IH, IB, OH, OB]
}

func (a *Api[P, Q, IH, IB, OH, OB]) Zeros() (P, Q, IH, IB) {
	var p P
	var q Q
	var reqHeaders IH
	var reqBody IB
	return p, q, reqHeaders, reqBody
}
