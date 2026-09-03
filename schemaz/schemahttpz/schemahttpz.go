package schemahttpz

import (
	"context"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type Req[Params any, Query any, ReqHeaders any, ReqBody any] struct {
	PathParams  Params
	QueryParams Query
	ReqHeaders  ReqHeaders
	ReqBody     ReqBody
}

// type Resp[RespHeaders any, RespBody any] struct {
// 	RespHeaders RespHeaders
// 	RespBody    RespBody
// }

type Handler[Params any, Query any, ReqHeaders any, ReqBody any, RespHeaders any, RespBody any] func(ctx context.Context, req *Req[Params, Query, ReqHeaders, ReqBody]) (RespHeaders, RespBody, error)

type Api[Params any, Query any, ReqHeaders any, ReqBody any, RespHeaders any, RespBody any] struct {
	Schema  *schemaz.Api
	Handler Handler[Params, Query, ReqHeaders, ReqBody, RespHeaders, RespBody]
}

func (a *Api[Params, Query, ReqHeaders, ReqBody, RespHeaders, RespBody]) Zeros() (Params, Query, ReqHeaders, ReqBody) {
	var p Params
	var q Query
	var reqHeaders ReqHeaders
	var reqBody ReqBody
	return p, q, reqHeaders, reqBody
}

func (a *Api[Params, Query, ReqHeaders, ReqBody, RespHeaders, RespBody]) ParseRequest(ctx context.Context, req *httpzserver.Req, params map[string]string) *Req[Params, Query, ReqHeaders, ReqBody] {
	p, q, reqHeaders, reqBody := a.Zeros()
	jsonz.Copy(params, &p)
	jsonz.Copy(req.Query, &q)
	jsonz.Copy(req.Headers, &reqHeaders)
	jsonz.Copy(req.Body, &reqBody)
	return &Req[Params, Query, ReqHeaders, ReqBody]{
		PathParams:  p,
		QueryParams: q,
		ReqHeaders:  reqHeaders,
		ReqBody:     reqBody,
	}
}

func Add[Params any, Query any, ReqHeaders any, ReqBody any, RespHeaders any, RespBody any](s *httpzserver.Server, api *Api[Params, Query, ReqHeaders, ReqBody, RespHeaders, RespBody]) {
	s.AddHandler(api.Schema.Method, api.Schema.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
		parsedReq := api.ParseRequest(ctx, req, params)
		respHedaers, respBody, err := api.Handler(ctx, parsedReq)
		if err != nil {
			panic(err)
		}
		var formattedHeaders http.Header
		jsonz.Copy(respHedaers, &formattedHeaders)
		w := resp(2000, formattedHeaders)
		jsonz.NewWriter[RespBody](w).MustWriteItem(respBody)
		// jsonz.Copy(respBody, &respBody)
		// resp(200, )
	})
}

// 	s.AddHandler(api.Method, api.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
// 		if api.ReqBody != nil {
// 			// reqBody := jsonz.NewReader[I](req.Body).MustReadItem()
// 		}

// 	})
// }
