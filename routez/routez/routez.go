package routez

import (
	"context"
	"net/http"
	"strings"
	"unicode"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/jsonz/structjsonz"
	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/infinity6-ai/gox/routez/apiz"
)

func header2json(headers http.Header) map[string][]string {
	n := make(map[string][]string, len(headers))
	for k, v := range headers {
		nk := strings.Map(func(r rune) rune {
			if r == '-' {
				return '_'
			}
			return unicode.ToLower(r)
		}, k)
		n[nk] = v
	}
	return n
}

func json2header(in map[string][]string, out http.Header) {
	for k, v := range in {
		nk := strings.ReplaceAll(k, "_", "-")
		out.Del(nk)
		for _, vv := range v {
			out.Add(nk, vv)
		}
	}
}

func parseRequest[P any, Q any, IH any, IB any, OH any, OB any](a *apiz.Api[P, Q, IH, IB, OH, OB], req *httpzserver.Req, params map[string]string) *apiz.Req[P, Q, IH, IB] {
	p, q, reqHeaders, reqBody := a.Zeros()
	structjsonz.MustParseSingle(params, &p)
	structjsonz.MustParse(req.Query, &q)
	structjsonz.MustParse(header2json(req.Headers), &reqHeaders)
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
		s.AddHandler(api.Schema.Method, api.Schema.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
			parsedReq := parseRequest(api, req, params)
			parsedResp, err := api.Handler(ctx, parsedReq)
			errorz.Check(err)
			formattedHeaders := make(http.Header)
			formattedHeaders.Set("Content-Type", "application/json")
			mapRespHedaers := structjsonz.MustFormat(parsedResp.RespHeaders)
			json2header(mapRespHedaers, formattedHeaders)
			w := resp(200, formattedHeaders)
			jsonz.FormatWriter(w, parsedResp.RespBody)
		})
	}
}
