package apiclientz

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/deferz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/routez/internal/converter"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

func MustSuccess(ctx context.Context, client *httpzclient.Client, reqResp apiz.Api) int {
	status := MustDo(ctx, client, reqResp)
	checker.GreaterOrEqual(status, 200, "http status error")
	checker.Less(status, 300, "http status erro")
	return status
}

func MustDo(ctx context.Context, client *httpzclient.Client, reqResp apiz.Api) int {
	status, err := Do(ctx, client, reqResp)
	errorz.Check(err)
	return status
}

func Do(ctx context.Context, client *httpzclient.Client, reqResp apiz.Api) (int, error) {
	nReq, closer, err := parseRequest(ctx, reqResp)
	if err != nil {
		return 0, err
	}
	defer closer.Close()
	nResp, err := client.Do(ctx, nReq)
	if err != nil {
		return 0, fmt.Errorf("%w: error calling server", err)
	}
	defer nResp.Body.Close()
	err = writeResponse(nResp, reqResp)
	return nResp.StatusCode, err
}

func writeResponse[T apiz.Api](nResp *httpzclient.Resp, api T) error {
	refs := api.ApiDataRefs()
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

type MyMap map[string][]string

func (s MyMap) UnmarshalJSON(data []byte) error {
	m := make(map[string]any)
	_, err := jsonz.Parse(data, &m)
	if err != nil {
		return fmt.Errorf("error parsing MyMap: %w", err)
	}
	for k, raw := range m {
		s[k] = nil
		if raw == nil {
			continue
		}
		values, ok := raw.([]any)
		if !ok {
			values = []any{raw}
		}
		nv := make([]string, len(values))
		s[k] = nv
		for idx, value := range values {
			switch v := value.(type) {
			case string:
				nv[idx] = v
			default:
				f, err := jsonz.Format(value)
				if err != nil {
					return fmt.Errorf("error parsing MyMap: %w", err)
				}
				nv[idx] = f.String()
			}
		}
	}
	return nil
}

func parseRequest[T apiz.Api](ctx context.Context, api T) (*httpzrequest.Req, io.Closer, error) {
	dfz := deferz.New(ctx)
	defer dfz.Close()
	refs := api.ApiDataRefs()
	var p map[string]string
	var h map[string][]string
	q := MyMap{}
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
	ret.Query = url.Values(q)
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
