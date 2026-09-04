package apiz

import (
	"context"

	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

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
