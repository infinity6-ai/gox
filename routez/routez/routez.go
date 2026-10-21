package routez

import (
	"context"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/routez/internal/converter"
)

func writeResponse(resp *apiz.Resp, serverResp httpzserver.Resp, refs *apiz.DataRefs) {
	if refs.RespHeaders != nil {
		headers := map[string][]string{}
		jsonz.MustCopy(refs.RespHeaders, &headers)
		converter.Json2Header(headers, resp.Headers)
	}
	w := serverResp(resp.Status, resp.Headers)
	jsonz.FormatWriter(w, refs.RespBody)
}

func parseRequest(refs *apiz.DataRefs, req *httpzrequest.Req, params map[string]string) {
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
}

func Register(s *httpzserver.Server, services ...apiz.Service) {
	for _, service := range services {
		spec := service.Api().ApiSpec()
		s.AddHandler(spec.Method, spec.Path, func(ctx context.Context, serverResp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			service := service.New()
			refs := service.Api().ApiDataRefs()
			parseRequest(refs, req, params)
			resp := service.Handler(ctx)
			if resp == nil {
				resp = &apiz.Resp{Status: 200}
			}
			if resp.Status == 0 {
				resp.Status = 200
			}
			if resp.Headers == nil {
				resp.Headers = make(http.Header)
			}
			if resp.Headers.Get("Content-Type") == "" {
				resp.Headers.Set("Content-Type", "application/json")
			}
			writeResponse(resp, serverResp, refs)
		})
	}
}
