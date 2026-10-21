package routez_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/routez/routez"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
	"github.com/stretchr/testify/require"
)

type dummyApi struct {
	spec apiz.ApiSpec
	refs apiz.DataRefs
}

func (a *dummyApi) ApiSpec() apiz.ApiSpec {
	return a.spec
}

func (a *dummyApi) ApiDataRefs() *apiz.DataRefs {
	return &a.refs
}

type dummyService struct {
	api     *dummyApi
	handler func(ctx context.Context) *apiz.Resp
}

func (s *dummyService) New() apiz.Service {
	return &dummyService{api: s.api, handler: s.handler}
}

func (s *dummyService) Api() apiz.Api {
	return s.api
}

func (s *dummyService) Handler(ctx context.Context) *apiz.Resp {
	return s.handler(ctx)
}

func TestUnitNilHandlerReturns200(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{})
	defer s.Close()
	s.Listen()
	s.Start()

	svc := &dummyService{
		api: &dummyApi{
			spec: apiz.ApiSpec{
				Id:     "test-nil",
				Method: "GET",
				Path:   "/test/nil",
			},
			refs: apiz.DataRefs{
				RespBody: &schemaz.Schema{
					Raw: func() any {
						return map[string]string{"status": "ok"}
					},
				},
			},
		},
		handler: func(ctx context.Context) *apiz.Resp {
			return nil
		},
	}

	routez.Register(s, svc)

	c := httpzclient.New(ctx, httpzclient.Options{
		BaseUrl: s.Base(),
	})

	req := httpzrequest.New("GET", "/test/nil")
	resp, err := c.Do(ctx, req)
	errorz.Check(err)
	defer resp.Body.Close()

	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "application/json", resp.Headers.Get("content-type"))
}

func TestUnitRespHeaders(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{})
	defer s.Close()
	s.Listen()
	s.Start()

	svc := &dummyService{
		api: &dummyApi{
			spec: apiz.ApiSpec{
				Id:     "test-headers",
				Method: "GET",
				Path:   "/test/headers",
			},
			refs: apiz.DataRefs{
				RespBody: &schemaz.Schema{
					Raw: func() any {
						return map[string]string{"status": "ok"}
					},
				},
			},
		},
		handler: func(ctx context.Context) *apiz.Resp {
			h := make(http.Header)
			h.Set("X-Custom-Header", "hello-world")
			return &apiz.Resp{
				Status:  201,
				Headers: h,
			}
		},
	}

	routez.Register(s, svc)

	c := httpzclient.New(ctx, httpzclient.Options{
		BaseUrl: s.Base(),
	})

	req := httpzrequest.New("GET", "/test/headers")
	resp, err := c.Do(ctx, req)
	errorz.Check(err)
	defer resp.Body.Close()

	require.Equal(t, 201, resp.StatusCode)
	require.Equal(t, "application/json", resp.Headers.Get("content-type"))
	require.Equal(t, "hello-world", resp.Headers.Get("x-custom-header"))
}
