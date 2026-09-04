package apiclientz

import (
	"bytes"
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/jsonz/structjsonz"
	"github.com/infinity6-ai/gox/httpz/client/httpzclient"
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/routez/internal/converter"
)

func parseRequest[P any, Q any, IH any, IB any, OH any, OB any](api *apiz.Api[P, Q, IH, IB, OH, OB], req *apiz.Req[P, Q, IH, IB]) (*httpzclient.Req, error) {
	p, err := structjsonz.FormatSingle(&req.PathParams)
	if err != nil {
		return nil, fmt.Errorf("%w: error formatting path params", err)
	}
	q, err := structjsonz.Format(&req.QueryParams)
	if err != nil {
		return nil, fmt.Errorf("%w: error formatting req query", err)
	}
	h, err := structjsonz.Format(&req.ReqHeaders)
	if err != nil {
		return nil, fmt.Errorf("%w: error formatting req headers", err)
	}

	ret, err := httpzclient.FormatReq(api.Schema.Method, api.Schema.Path, p)
	if err != nil {
		return nil, fmt.Errorf("%w: error formatting request", err)
	}
	ret.Query = q
	converter.Json2Header(h, ret.Headers)
	fBody, err := jsonz.Format(req.ReqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: error formatting request body", err)
	}
	ret.Body = bytes.NewReader(fBody.Bytes())
	return ret, nil
}

func Get[P any, Q any, IH any, IB any, OH any, OB any](client *httpzclient.Client, api *apiz.Api[P, Q, IH, IB, OH, OB]) apiz.Handler[P, Q, IH, IB, OH, OB] {
	return func(ctx context.Context, req *apiz.Req[P, Q, IH, IB]) (*apiz.Resp[OH, OB], error) {

		nReq, err := parseRequest(api, req)
		if err != nil {
			return nil, err
		}
		nResp, err := client.Do(ctx, nReq)
		if err != nil {
			return nil, fmt.Errorf("%w: error calling server", err)
		}
		defer nResp.Body.Close()
		ret := &apiz.Resp[OH, OB]{
			Status: nResp.StatusCode,
		}
		convertedHeaders := converter.Header2Json(nResp.Headers)
		structjsonz.Parse(convertedHeaders, &ret.RespHeaders)
		_, err = jsonz.ParseReader(nResp.Body, &ret.RespBody)
		if err != nil {
			return nil, fmt.Errorf("%w: error parsing response body", err)
		}
		return ret, nil
	}
}
