package httpzserverfilterlogz

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/logz/logzctx"
	"github.com/infinity6-ai/gox/commonz/logz/tracerz"
	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const logzCtxKeyReqId = "http.server.req.id"

// const LogzCtxKeyServerName = "http.server.name"
// const LogzCtxKeyServerVersion = "http.server.version"

// var I6_LOGZ_FILTER_GAE = configz.Create("I6_LOGZ_FILTER_GAE", "false")

func Filter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	reqId := req.Headers.Get(httpz.HeaderReqId)
	ctx = logz.SetAttr(ctx, logzCtxKeyReqId, reqId)

	tracerInfoHeader := propagation.HeaderCarrier(req.Headers)
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, tracerInfoHeader)

	m := logzctx.Get(ctx)
	appName := m[logz.CtxKeyAppName]
	appVersion := m[logz.CtxKeyAppVersion]

	ctx, span := tracerz.Span(ctx, "httpz.server", appName, req.Method, req.Path.String())
	defer span.Close()

	tracerz.AttrString(ctx, "req.id", reqId)
	tracerz.AttrString(ctx, logz.CtxKeyAppName, appName)
	tracerz.AttrString(ctx, logz.CtxKeyAppVersion, appVersion)

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
