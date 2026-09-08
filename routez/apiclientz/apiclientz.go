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
