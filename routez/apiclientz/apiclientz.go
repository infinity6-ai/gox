package apiclientz

import (
	"bytes"
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/jsonz/structjsonz"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/routez/internal/converter"
)

func parseRequest[T apiz.ReqResp](api *apiz.Api[T], reqResp T) (*httpzrequest.Req, error) {
	refs := reqResp.GetDataRefs()
	p, err := structjsonz.FormatSingle(refs.PathParams)
	if err != nil {
		return nil, fmt.Errorf("%w: error formatting path params", err)
	}
	q, err := structjsonz.Format(refs.QueryParams)
	if err != nil {
		return nil, fmt.Errorf("%w: error formatting req query", err)
	}
	h, err := structjsonz.Format(refs.ReqHeaders)
	if err != nil {
		return nil, fmt.Errorf("%w: error formatting req headers", err)
	}

	ret, err := httpzrequest.Format(api.Schema.Method, api.Schema.Path, p)
	if err != nil {
		return nil, fmt.Errorf("%w: error formatting request", err)
	}
	ret.Query = q
	converter.Json2Header(h, ret.Headers)
	fBody, err := jsonz.Format(refs.ReqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: error formatting request body", err)
	}
	ret.Body = bytes.NewReader(fBody.Bytes())
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
		writeResponse(nResp, reqResp)
		return nResp.StatusCode, nil
	}
}

func writeResponse[T apiz.ReqResp](nResp *httpzclient.Resp, reqResp T) {
	refs := reqResp.GetDataRefs()
	convertedHeaders := converter.Header2Json(nResp.Headers)
	structjsonz.Parse(convertedHeaders, refs.RespHeaders)
	_, err := jsonz.ParseReader(nResp.Body, refs.RespBody)
	errorz.Check(err)
}
