package httpzserverfilterlogz

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/configz"
	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/logz/logzctx"
	"github.com/infinity6-ai/gox/commonz/logz/logzprovider/logzproviderstructured"
	"github.com/infinity6-ai/gox/commonz/logz/tracerz"
	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const LOGZ_CTX_REQ_ID = "http.server.req.id"
const LOGZ_CTX_SERVER_NAME = "http.server.name"
const LOGZ_CTX_SERVER_VERSION = "http.server.version"

var I6_LOGZ_FILTER_GAE = configz.Create("I6_LOGZ_FILTER_GAE", "false")

type LogzFilter struct {
	ServName  string
	ProjectId string
	Attrs     map[string]string
}

func (me *LogzFilter) Filter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	reqId := req.Headers.Get(httpz.HeaderReqId)
	m := map[string]string{
		LOGZ_CTX_REQ_ID: reqId,
	}
	m[logzproviderstructured.CtxKeyGaeProjectName] = me.ProjectId
	ctx = logz.SetAttrs(ctx, m)

	tracerInfoHeader := propagation.HeaderCarrier(req.Headers)
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, tracerInfoHeader)

	m = logzctx.Get(ctx)
	appName := m[logz.CtxKeyAppName]

	ctx, span := tracerz.Span(ctx, "httpz.server", appName, req.Method, req.Path.String())
	defer span.Close()

	for k, v := range me.Attrs {
		tracerz.AttrString(ctx, k, v)
	}
	tracerz.AttrString(ctx, "req.id", reqId)

	next(ctx, func(status int, headers http.Header) io.Writer {
		tracerz.AttrInt(ctx, "resp.status_code", status)

		if status >= 200 && status < 300 {
			tracerz.Ok(ctx, fmt.Sprintf("%d", status))
		}

		if status >= 400 {
			tracerz.Error(ctx, fmt.Sprintf("%d", status))
		}

		return resp(status, headers)
	}, req)
}
