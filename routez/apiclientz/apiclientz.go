package apiclientz

import (
	"bytes"
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/jsonz/structjsonz"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/routez/internal/converter"
	"github.com/infinity6-ai/gox/schemaz/schemazv2"
)

func parseRequest[T apiz.ReqResp](api *apiz.Api[T], reqResp T) (*httpzrequest.Req, error) {
	refs := reqResp.GetDataRefs()
	var p map[string]string
	var q, h map[string][]string
	var err error
	if refs.PathParams != nil {
		p, err = structjsonz.FormatSingle(refs.PathParams)
		if err != nil {
			return nil, fmt.Errorf("%w: error formatting path params", err)
		}
	}
	if refs.QueryParams != nil {
		q, err = structjsonz.Format(refs.QueryParams)
		if err != nil {
			return nil, fmt.Errorf("%w: error formatting req query", err)
		}
	}
	if refs.ReqHeaders != nil {
		h, err = structjsonz.Format(refs.ReqHeaders)
		if err != nil {
			return nil, fmt.Errorf("%w: error formatting req headers", err)
		}
	}
	ret, err := httpzrequest.Format(api.Schema.Method, api.Schema.Path, p)
	if err != nil {
		return nil, fmt.Errorf("%w: error formatting request", err)
	}
	ret.Query = q
	converter.Json2Header(h, ret.Headers)
	if refs.ReqBody != nil {
		fBody, err := jsonz.Format(refs.ReqBody)
		if err != nil {
			return nil, fmt.Errorf("%w: error formatting request body", err)
		}
		ret.Body = bytes.NewReader(fBody.Bytes())
	}
	return ret, nil
}

func Get[T apiz.ReqResp](client *httpzclient.Client, api *apiz.Api[T]) apiz.Handler[T] {
	return func(ctx context.Context, reqResp T) (int, error) {
		nReq, err := parseRequest(api, reqResp)
		if err != nil {
			return 0, err
		}
		nResp, err := client.Do(ctx, nReq)
		if err != nil {
			return 0, fmt.Errorf("%w: error calling server", err)
		}
		defer nResp.Body.Close()
		err = writeResponse(nResp, reqResp)
		return nResp.StatusCode, err
	}
}

func GetV2[T apiz.ReqRespV2](client *httpzclient.Client, api *apiz.ApiV2[T]) apiz.HandlerV2[T] {
	return func(ctx context.Context, reqResp T) (int, error) {
		nReq, err := parseRequestV2(api, reqResp)
		if err != nil {
			return 0, err
		}
		nResp, err := client.Do(ctx, nReq)
		if err != nil {
			return 0, fmt.Errorf("%w: error calling server", err)
		}
		defer nResp.Body.Close()
		err = writeResponseV2(nResp, reqResp)
		return nResp.StatusCode, err
	}
}

func writeResponseV2[T apiz.ReqRespV2](nResp *httpzclient.Resp, reqResp T) error {
	refs := reqResp.GetDataRefsV2()
	if refs.RespHeaders != nil {
		convertedHeaders := converter.Header2Json(nResp.Headers)
		err := structjsonz.Parse(convertedHeaders, refs.RespHeaders)
		if err != nil {
			return fmt.Errorf("%w: error parsing resp headers", err)
		}
	}
	if refs.RespBody != nil {
		_, err := jsonz.ParseReader(nResp.Body, refs.RespBody)
		if err != nil {
			return fmt.Errorf("%w: error parsing resp body", err)
		}
	}
	return nil
}

func parseRequestV2[T apiz.ReqRespV2](api *apiz.ApiV2[T], reqResp T) (*httpzrequest.Req, error) {
	refs := reqResp.GetDataRefsV2()
	var p map[string]string
	// var q, h map[string][]string
	// var err error
	if refs.PathParams != nil {
		err := formatPathParams(refs.PathParams, &p)
		if err != nil {
			return nil, fmt.Errorf("%w: error formatting path params", err)
		}
	}
	// if refs.QueryParams != nil {
	// 	q, err = structjsonz.Format(refs.QueryParams)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("%w: error formatting req query", err)
	// 	}
	// }
	// if refs.ReqHeaders != nil {
	// 	h, err = structjsonz.Format(refs.ReqHeaders)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("%w: error formatting req headers", err)
	// 	}
	// }
	// ret, err := httpzrequest.Format(api.Schema.Method, api.Schema.Path, p)
	// if err != nil {
	// 	return nil, fmt.Errorf("%w: error formatting request", err)
	// }
	// ret.Query = q
	// converter.Json2Header(h, ret.Headers)
	// if refs.ReqBody != nil {
	// 	fBody, err := jsonz.Format(refs.ReqBody)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("%w: error formatting request body", err)
	// 	}
	// 	ret.Body = bytes.NewReader(fBody.Bytes())
	// }
	// return ret, nil
	panic("xxx")
}

func formatPathParams(schema *schemazv2.Schema, out *map[string]string) error {
	var m map[string][]string
	_, err := jsonz.Copy(schema, &m)
	if err != nil {
		return err
	}
	*out = converter.Json2Params(m)
	return nil
}

func writeResponse[T apiz.ReqResp](nResp *httpzclient.Resp, reqResp T) error {
	refs := reqResp.GetDataRefs()
	if refs.RespHeaders != nil {
		convertedHeaders := converter.Header2Json(nResp.Headers)
		err := structjsonz.Parse(convertedHeaders, refs.RespHeaders)
		if err != nil {
			return fmt.Errorf("%w: error parsing resp headers", err)
		}
	}
	if refs.RespBody != nil {
		_, err := jsonz.ParseReader(nResp.Body, refs.RespBody)
		if err != nil {
			return fmt.Errorf("%w: error parsing resp body", err)
		}
	}
	return nil
}
