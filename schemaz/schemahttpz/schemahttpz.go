package schemahttpz

import (
	"context"

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

func Add[Params any, Query any, ReqHeaders any, ReqBody any, RespHeaders any, RespBody any](s *httpzserver.Server, api *Api[Params, Query, ReqHeaders, ReqBody, RespHeaders, RespBody]) {

}

// 	s.AddHandler(api.Method, api.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
// 		if api.ReqBody != nil {
// 			// reqBody := jsonz.NewReader[I](req.Body).MustReadItem()
// 		}

// 	})
// }
