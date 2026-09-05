package routez

import (
	"context"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/jsonz/structjsonz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/routez/internal/converter"
)

func parseRequest[P any, Q any, IH any, IB any, OH any, OB any](a *apiz.Api[P, Q, IH, IB, OH, OB], req *httpzrequest.Req, params map[string]string) *apiz.Req[P, Q, IH, IB] {
	p, q, reqHeaders, reqBody := a.Zeros()
	structjsonz.MustParseSingle(params, &p)
	structjsonz.MustParse(req.Query, &q)
	structjsonz.MustParse(converter.Header2Json(req.Headers), &reqHeaders)
	jsonz.MustParseReader(req.Body, &reqBody)
	return &apiz.Req[P, Q, IH, IB]{
		PathParams:  p,
		QueryParams: q,
		ReqHeaders:  reqHeaders,
		ReqBody:     reqBody,
	}
}

func Register[P any, Q any, IH any, IB any, OH any, OB any](s *httpzserver.Server, apis ...*apiz.Api[P, Q, IH, IB, OH, OB]) {
	for _, api := range apis {
		s.AddHandler(api.Schema.Method, api.Schema.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			parsedReq := parseRequest(api, req, params)
			parsedResp, err := api.Handler(ctx, parsedReq)
			errorz.Check(err)
			formattedHeaders := make(http.Header)
			formattedHeaders.Set("Content-Type", "application/json")
			mapRespHedaers := structjsonz.MustFormat(&parsedResp.RespHeaders)
			converter.Json2Header(mapRespHedaers, formattedHeaders)
			w := resp(parsedResp.Status, formattedHeaders)
			jsonz.FormatWriter(w, parsedResp.RespBody)
		})
	}
}

func parseRequestV2[T apiz.ReqResp](a *apiz.ApiV2[T], req *httpzrequest.Req, params map[string]string) T {
	reqResp := a.MewReqResp()
	refs := reqResp.GetDataRefs()
	structjsonz.MustParseSingle(params, refs.PathParams)
	structjsonz.MustParse(req.Query, refs.QueryParams)
	structjsonz.MustParse(converter.Header2Json(req.Headers), refs.ReqHeaders)
	jsonz.MustParseReader(req.Body, refs.ReqBody)
	return reqResp
}

func writeResponse[T apiz.ReqResp](status int, resp httpzserver.Resp, reqResp T, formattedHeaders http.Header) {
	refs := reqResp.GetDataRefs()
	mapRespHedaers := structjsonz.MustFormat(refs.RespHeaders)
	converter.Json2Header(mapRespHedaers, formattedHeaders)
	w := resp(status, formattedHeaders)
	jsonz.FormatWriter(w, refs.RespBody)
}

func RegisterV2[T apiz.ReqResp](s *httpzserver.Server, apis ...*apiz.ApiV2[T]) {
	for _, api := range apis {
		s.AddHandler(api.Schema.Method, api.Schema.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			reqResp := parseRequestV2(api, req, params)
			status, err := api.HandlerV2(ctx, reqResp)
			errorz.Check(err)
			formattedHeaders := make(http.Header)
			formattedHeaders.Set("Content-Type", "application/json")
			writeResponse(status, resp, reqResp, formattedHeaders)
		})
	}
}
