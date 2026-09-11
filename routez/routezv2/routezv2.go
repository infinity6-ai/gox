package routezv2

import (
	"context"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/routez/apizv2"
	"github.com/infinity6-ai/gox/routez/internal/converter"
)

func writeResponse(status int, resp httpzserver.Resp, refs *apizv2.DataRefs, formattedHeaders http.Header) {
	if refs.RespHeaders != nil {
		headers := map[string][]string{}
		jsonz.MustCopy(refs.RespHeaders, &headers)
		converter.Json2Header(headers, formattedHeaders)
	}
	w := resp(status, formattedHeaders)
	jsonz.FormatWriter(w, refs.RespBody)
}

func parseRequest(refs *apizv2.DataRefs, req *httpzrequest.Req, params map[string]string) {
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

func Register(s *httpzserver.Server, services ...apizv2.Service) {
	for _, service := range services {
		spec := service.Api().ApiSpec()
		s.AddHandler(spec.Method, spec.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			service := service.New()
			refs := service.Api().GetDataRefs()
			parseRequest(refs, req, params)
			status, err := service.Handler(ctx)
			errorz.Check(err)
			formattedHeaders := make(http.Header)
			formattedHeaders.Set("Content-Type", "application/json")
			writeResponse(status, resp, refs, formattedHeaders)
		})
	}
}
