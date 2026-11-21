package jsapiz

import (
	"context"
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/stringz/camelz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/schemaz/jsschemaz"
	"github.com/infinity6-ai/gox/schemaz/typeastz"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type Req struct {
	Params  typeastz.Typedef `json:"params"`
	Query   typeastz.Typedef `json:"query"`
	Headers typeastz.Typedef `json:"headers"`
	Payload typeastz.Typedef `json:"payload"`
}

type Resp struct {
	Headers typeastz.Typedef `json:"headers"`
	Payload typeastz.Typedef `json:"payload"`
}

type Api struct {
	Spec     ApiSpec          `json:"spec"`
	Typedefs typeastz.Typedefs `json:"typedefs"`

	Req  Req  `json:"req"`
	Resp Resp `json:"resp"`
}

type ApiSpec struct {
	Id      string        `json:"id,omitempty"`
	CamelId string        `json:"camel_id"`
	Desc    *schemaz.Desc `json:"desc,omitempty"`
	Method  string        `json:"method,omitempty"`
	Path    string        `json:"path,omitempty"`
}

type Apis struct {
	Apis     []*Api           `json:"apis"`
	Typedefs typeastz.Typedefs `json:"typedefs"`
}

type Parser struct {
	schemaParser *jsschemaz.Parser
}

func New(schemaParser *jsschemaz.Parser) *Parser {
	if schemaParser == nil {
		schemaParser = jsschemaz.New()
	}
	return &Parser{schemaParser: schemaParser}
}

func (p *Parser) ParseApi(ctx context.Context, api apiz.Api) *Api {
	refs := api.ApiDataRefs()

	apiId := ""
	if id := api.ApiSpec().Id; id != "" {
		apiId = "api" + camelz.P(id).P()
	}

	parseSpec := func() ApiSpec {
		s := api.ApiSpec()
		ret := ApiSpec{
			Id:      s.Id,
			CamelId: apiId,
			Method:  s.Method,
			Path:    s.Path,
		}
		if s.Desc != nil {
			ret.Desc = s.Desc()
		}
		return ret
	}

	ret := &Api{
		Spec: parseSpec(),
	}

	xxx := typeastz.BasicTypes()
	ret.Typedefs.ExtractFrom(xxx)

	validateDataRef(refs.PathParams, "PathParams")
	validateDataRef(refs.QueryParams, "QueryParams")
	validateDataRef(refs.ReqHeaders, "ReqHeaders")
	validateDataRef(refs.RespHeaders, "RespHeaders")

	parseRefsV2 := func(out *typeastz.Typedef, suffix string, s *schemaz.Schema) {
		if s == nil {
			return
		}
		fieldName := ret.Spec.CamelId + strings.ToUpper(suffix[0:1]) + suffix[1:]
		td, tdfs := p.schemaParser.ParseSchema(ctx, fieldName, s)
		checker.Equal("object", td.Head.Typ, "must be object")
		out.Properties = &typeastz.Tuples{}
		out.Head = &typeastz.Tuple{Typ: fieldName, Name: suffix}
		td.Head.Summary = fmt.Sprintf("%s %s", ret.Spec.CamelId, suffix)
		ret.Typedefs.ExtractFrom(tdfs)
		ret.Typedefs.Add(td)
	}

	parseRefsV2(&ret.Req.Params, "params", refs.PathParams)
	parseRefsV2(&ret.Req.Query, "query", refs.QueryParams)
	parseRefsV2(&ret.Req.Headers, "reqHeaders", refs.ReqHeaders)
	parseRefsV2(&ret.Req.Payload, "reqPayload", refs.ReqBody)
	parseRefsV2(&ret.Resp.Headers, "respHeaders", refs.RespHeaders)
	parseRefsV2(&ret.Resp.Payload, "respPayload", refs.RespBody)

	return ret
}

func (p *Parser) ParseApis(ctx context.Context, apis []apiz.Api) *Apis {
	ret := Apis{
		Apis:     make([]*Api, len(apis)),
		Typedefs: typeastz.NewTypedefs(nil),
	}
	for idx, api := range apis {
		parsed := p.ParseApi(ctx, api)
		ret.Typedefs.Merge(parsed.Typedefs)
		ret.Apis[idx] = parsed
	}
	return &ret
}

func ExtractTypedefs(apis []*Api) typeastz.Typedefs {
	tdfs := typeastz.Typedefs{}
	for _, api := range apis {
		apiId := api.Spec.CamelId
		reqTypedef, hasReq := buildReqTypedef(apiId, api)
		respTypedef := buildRespTypedef(apiId, api)

		if hasReq {
			tdfs.Add(&reqTypedef)
		}
		tdfs.Add(&respTypedef)
	}
	return tdfs
}

func buildReqTypedef(apiId string, api *Api) (typeastz.Typedef, bool) {
	hasParams := api.Req.Params.Head != nil
	hasQuery := api.Req.Query.Head != nil
	hasReqHeaders := api.Req.Headers.Head != nil
	hasPayload := api.Req.Payload.Head != nil

	hasReq := hasParams || hasQuery || hasReqHeaders || hasPayload

	reqTypedef := typeastz.Typedef{
		Head:       &typeastz.Tuple{Typ: "object", Name: apiId + "Req", Summary: apiId + " Request"},
		Properties: &typeastz.Tuples{},
	}
	if hasParams {
		reqTypedef.Properties.Add(typeastz.Tuple{Typ: api.Req.Params.Head.Typ, GoTyp: api.Req.Params.Head.GoTyp, Name: "params"})
	}
	if hasQuery {
		reqTypedef.Properties.Add(typeastz.Tuple{Typ: api.Req.Query.Head.Typ, GoTyp: api.Req.Query.Head.GoTyp, Name: "query"})
	}
	if hasReqHeaders {
		reqTypedef.Properties.Add(typeastz.Tuple{Typ: api.Req.Headers.Head.Typ, GoTyp: api.Req.Headers.Head.GoTyp, Name: "headers"})
	}
	if hasPayload {
		reqTypedef.Properties.Add(typeastz.Tuple{Typ: api.Req.Payload.Head.Typ, GoTyp: api.Req.Payload.Head.GoTyp, Name: "payload"})
	}
	return reqTypedef, hasReq
}

func buildRespTypedef(apiId string, api *Api) typeastz.Typedef {
	respTypedef := typeastz.Typedef{
		Head:       &typeastz.Tuple{Typ: "object", Name: apiId + "Resp", Summary: apiId + " Response"},
		Properties: &typeastz.Tuples{},
	}
	if api.Resp.Headers.Head != nil {
		respTypedef.Properties.Add(typeastz.Tuple{Typ: api.Resp.Headers.Head.Typ, GoTyp: api.Resp.Headers.Head.GoTyp, Name: "headers"})
	}
	if api.Resp.Payload.Head != nil {
		respTypedef.Properties.Add(typeastz.Tuple{Typ: api.Resp.Payload.Head.Typ, GoTyp: api.Resp.Payload.Head.GoTyp, Name: "payload"})
	}
	return respTypedef
}

func validateDataRef(schema *schemaz.Schema, n string) {
	if schema != nil {
		checker.NotNil(schema.Object, "%s", n)
	}
}
