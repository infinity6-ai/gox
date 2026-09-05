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

func Register[T apiz.ReqResp](s *httpzserver.Server, apis ...*apiz.Api[T]) {
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
