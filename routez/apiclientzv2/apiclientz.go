package apiclientzv2

import (
	"context"
	"fmt"
	"io"

	"github.com/infinity6-ai/gox/commonz/deferz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/routez/apizv2"
	"github.com/infinity6-ai/gox/routez/internal/converter"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

func Get[T apizv2.Api](client *httpzclient.Client, api T) apizv2.Handler[T] {
	return func(ctx context.Context, api T) (int, error) {
		nReq, closer, err := parseRequest(ctx, api)
		if err != nil {
			return 0, err
		}
		defer closer.Close()
		nResp, err := client.Do(ctx, nReq)
		if err != nil {
			return 0, fmt.Errorf("%w: error calling server", err)
		}
		defer nResp.Body.Close()
		err = writeResponse(nResp, api)
		return nResp.StatusCode, err
	}
}

func writeResponse[T apizv2.Api](nResp *httpzclient.Resp, api T) error {
	refs := api.GetDataRefs()
	if refs.RespHeaders != nil {
		convertedHeaders := converter.Header2Json(nResp.Headers)
		_, err := jsonz.Copy(convertedHeaders, refs.RespHeaders)
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

func parseRequest[T apizv2.Api](ctx context.Context, api T) (*httpzrequest.Req, io.Closer, error) {
	dfz := deferz.New(ctx)
	defer dfz.Close()
	refs := api.GetDataRefs()
	var p map[string]string
	var q, h map[string][]string
	if refs.PathParams != nil {
		err := formatPathParams(refs.PathParams, &p)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: error formatting path params", err)
		}
	}
	if refs.QueryParams != nil {
		_, err := jsonz.Copy(refs.QueryParams, &q)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: error formatting req query", err)
		}
	}
	if refs.ReqHeaders != nil {
		_, err := jsonz.Copy(refs.ReqHeaders, &h)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: error formatting req headers", err)
		}
	}
	ret, err := httpzrequest.Format(api.ApiSpec().Method, api.ApiSpec().Path, p)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: error formatting request", err)
	}
	ret.Query = q
	converter.Json2Header(h, ret.Headers)
	if refs.ReqBody != nil {
		r := jsonz.FormatReadCloser(refs.ReqBody)
		dfz.AddCloserS(r)
		ret.Body = r
	}
	return ret, dfz.Detach(), nil
}

func formatPathParams(schema *schemaz.Schema, out *map[string]string) error {
	var m map[string][]string
	_, err := jsonz.Copy(schema, &m)
	if err != nil {
		return err
	}
	*out = converter.Json2Params(m)
	return nil
}
