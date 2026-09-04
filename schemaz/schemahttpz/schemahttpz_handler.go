package schemahttpz

import (
	"context"
	"net/http"
	"strings"
	"unicode"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/jsonz/structjsonz"
	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type Req[P any, Q any, IH any, IB any] struct {
	PathParams  P
	QueryParams Q
	ReqHeaders  IH
	ReqBody     IB
}

type Resp[OH any, OB any] struct {
	Status      int
	RespHeaders OH
	RespBody    OB
}

type Handler[P any, Q any, IH any, IB any, OH any, OB any] func(ctx context.Context, req *Req[P, Q, IH, IB]) (*Resp[OH, OB], error)

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

func (a *Api[P, Q, IH, IB, OH, OB]) ParseRequest(ctx context.Context, req *httpzserver.Req, params map[string]string) *Req[P, Q, IH, IB] {
	p, q, reqHeaders, reqBody := a.Zeros()
	structjsonz.MustParseSingle(params, &p)
	structjsonz.MustParse(req.Query, &q)
	structjsonz.MustParse(header2json(req.Headers), &reqHeaders)
	jsonz.MustParseReader(req.Body, &reqBody)
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
