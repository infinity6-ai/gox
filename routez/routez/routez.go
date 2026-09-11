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

func parseRequest[T apiz.ReqResp](a *apiz.Api[T], req *httpzrequest.Req, params map[string]string) T {
	reqResp := a.MewReqResp()
	refs := reqResp.GetDataRefs()
	if refs.PathParams != nil {
		structjsonz.MustParseSingle(params, refs.PathParams)
	}
	if refs.QueryParams != nil {
		structjsonz.MustParse(req.Query, refs.QueryParams)
	}
	if refs.ReqHeaders != nil {
		structjsonz.MustParse(converter.Header2Json(req.Headers), refs.ReqHeaders)
	}
	if refs.ReqBody != nil {
		jsonz.MustParseReader(req.Body, refs.ReqBody)
	}
	return reqResp
}

func writeResponseV2[T apiz.ReqRespV2](status int, resp httpzserver.Resp, reqResp T, formattedHeaders http.Header) {
	refs := reqResp.GetDataRefsV2()
	if refs.RespHeaders != nil {
		headers := map[string][]string{}
		jsonz.MustCopy(refs.ReqHeaders, &headers)
		converter.Json2Header(headers, formattedHeaders)
	}
	w := resp(status, formattedHeaders)
	jsonz.FormatWriter(w, refs.RespBody)
}

func writeResponse[T apiz.ReqResp](status int, resp httpzserver.Resp, reqResp T, formattedHeaders http.Header) {
	refs := reqResp.GetDataRefs()
	if refs.RespHeaders != nil {
		mapRespHedaers := structjsonz.MustFormat(refs.RespHeaders)
		converter.Json2Header(mapRespHedaers, formattedHeaders)
	}
	w := resp(status, formattedHeaders)
	jsonz.FormatWriter(w, refs.RespBody)
}

func RegisterOLD[T apiz.ReqResp](s *httpzserver.Server, apis ...*apiz.Api[T]) {
	for _, api := range apis {
		s.AddHandler(api.Schema.Method, api.Schema.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			reqResp := parseRequest(api, req, params)
			status, err := api.Handler(ctx, reqResp)
			errorz.Check(err)
			formattedHeaders := make(http.Header)
			formattedHeaders.Set("Content-Type", "application/json")
			writeResponse(status, resp, reqResp, formattedHeaders)
		})
	}
}

func parseRequestV2[T apiz.ReqRespV2](a *apiz.ApiV2[T], req *httpzrequest.Req, params map[string]string) T {
	reqResp := a.MewReqResp()
	refs := reqResp.GetDataRefsV2()
	if refs.PathParams != nil {
		jsonz.MustCopy(converter.Params2Json(params), refs.PathParams)
	}
	if refs.QueryParams != nil {
		jsonz.MustParse(jsonz.MustFormat(req.Query).Bytes(), refs.QueryParams)
	}
	if refs.ReqHeaders != nil {
		jsonz.MustParse(jsonz.MustFormat(converter.Header2Json(req.Headers)).Bytes(), refs.ReqHeaders)
	}
	if refs.ReqBody != nil {
		jsonz.MustParseReader(req.Body, refs.ReqBody)
	}
	return reqResp
}

func Register[T apiz.ReqRespV2](s *httpzserver.Server, apis ...*apiz.ApiV2[T]) {
	for _, api := range apis {
		s.AddHandler(api.Method, api.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			reqResp := parseRequestV2(api, req, params)
			status, err := api.Handler(ctx, reqResp)
			errorz.Check(err)
			formattedHeaders := make(http.Header)
			formattedHeaders.Set("Content-Type", "application/json")
			writeResponseV2(status, resp, reqResp, formattedHeaders)
		})
	}
}
