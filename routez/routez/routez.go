package routez

import (
	"context"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/routez/internal/converter"
)

func writeResponse[T apiz.Api](status int, resp httpzserver.Resp, reqResp T, formattedHeaders http.Header) {
	refs := reqResp.GetDataRefs()
	if refs.RespHeaders != nil {
		headers := map[string][]string{}
		jsonz.MustCopy(refs.RespHeaders, &headers)
		converter.Json2Header(headers, formattedHeaders)
	}
	w := resp(status, formattedHeaders)
	jsonz.FormatWriter(w, refs.RespBody)
}

func parseRequest[T apiz.Api](a *apiz.Spec[T], req *httpzrequest.Req, params map[string]string) T {
	reqResp := a.MewReqResp()
	refs := reqResp.GetDataRefs()
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

func Register[T apiz.Api](s *httpzserver.Server, services ...*apiz.Service[T]) {
	for _, service := range services {
		s.AddHandler(service.Spec.Method, service.Spec.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			reqResp := parseRequest(service.Spec, req, params)
			status, err := service.Handler(ctx, reqResp)
			errorz.Check(err)
			formattedHeaders := make(http.Header)
			formattedHeaders.Set("Content-Type", "application/json")
			writeResponse(status, resp, reqResp, formattedHeaders)
		})
	}
}
