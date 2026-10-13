// Package httpzserverfiltererror provides panic recovery and error handling middleware
// for httpzserver, formatting panics as structured errorz JSON responses and logging them.
package httpzserverfiltererror

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

// Collector returns the log collector for capturing and asserting error entries in tests.
func Collector() func() []*logzspec.Entry {
	return logger.(logz.Collector).Collector()
}

// Filter recovers panics in downstream handlers, logs them, and formats structured error responses.
// If the response was already started when the panic occurred, it logs and re-panics with fmt.Errorf.
func Filter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	responseStarted := false

	nResp := func(status int, headers http.Header) io.Writer {
		responseStarted = true
		return resp(status, headers)
	}

	defer func() {
		r := recover()
		if r == nil {
			return
		}
		stack := debug.Stack()
		formattedStack := string(stack)
		formattedError := fmt.Sprintf("%#v", r)
		errError, _ := r.(error)
		if errError == nil {
			errError = fmt.Errorf("%v", r)
		}
		logger.Error(ctx, "http handler panic", map[string]any{"err": formattedError, "stack": formattedStack, "resp_started": responseStarted}, errError)

		if responseStarted {
			panic("we got a panic, but response was already started")
		}

		code, payload := formatErrorz(r)
		h := make(http.Header)
		h.Add(httpz.HeaderError, "true")
		h.Set("Content-Type", "application/json")
		w := resp(code, h)
		_ = jsonz.FormatWriter(w, payload)
	}()

	next(ctx, nResp, req)
}

func formatErrorz(err any) (code int, payload any) {
	errz := errorz.From(err)
	code = http.StatusInternalServerError
	payload = errz.Internal()
	if errz.Business {
		code = errz.Code
		if code == 0 {
			code = http.StatusBadRequest
			errz.Code = code
		}
		payload = errz
	}
	return
}
