package schemahttpz

import (
	"context"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type Req[P any, Q any, IH any, IB any] struct {
	PathParams  P
	QueryParams Q
	ReqHeaders  IH
	ReqBody     IB
}

type Handler[P any, Q any, IH any, IB any, OH any, OB any] func(ctx context.Context, req *Req[P, Q, IH, IB]) (OH, OB, error)

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

func (a *Api[P, Q, IH, IB, OH, OB]) ParseRequest(ctx context.Context, req *httpzserver.Req, params map[string]string) *Req[P, Q, IH, IB] {
	p, q, reqHeaders, reqBody := a.Zeros()
	jsonz.Copy(params, &p)
	jsonz.Copy(req.Query, &q)
	jsonz.Copy(req.Headers, &reqHeaders)
	jsonz.Copy(req.Body, &reqBody)
	return &Req[P, Q, IH, IB]{
		PathParams:  p,
		QueryParams: q,
		ReqHeaders:  reqHeaders,
		ReqBody:     reqBody,
	}
}

func Add[P any, Q any, IH any, IB any, OH any, OB any](s *httpzserver.Server, api *Api[P, Q, IH, IB, OH, OB]) {
	s.AddHandler(api.Schema.Method, api.Schema.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
		parsedReq := api.ParseRequest(ctx, req, params)
		respHedaers, _, err := api.Handler(ctx, parsedReq)
		errorz.Check(err)
		formattedHeaders := make(http.Header)
		formattedHeaders.Set("Content-Type", "application/json")
		jsonz.Copy(respHedaers, &formattedHeaders)
		w := resp(200, formattedHeaders)
		jsonz.NewWriter[*OB](w).MustWriteItem(nil)
	})
}
