package httpzserverfiltercorz

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

type CorzFilter struct {
	mu       sync.Mutex
	Methods  []string
	Headers  http.Header
	patterns []string
	prefixes []string
}

func NewCorzFilter() *CorzFilter {
	return newCorzFilter()
}

func newCorzFilter() *CorzFilter {
	f := &CorzFilter{}
	f.fix()
	return f
}

func (me *CorzFilter) RestPattern(pattern string) {
	me.patterns = append(me.patterns, pattern)
}

func (me *CorzFilter) RestPrefix(prefix string) {
	me.prefixes = append(me.prefixes, prefix)
}

func (me *CorzFilter) Filter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	me.fix()

	// pathStr := ""
	// if req != nil && req.Path != nil {
	// 	pathStr = req.Path.String()
	// }

	// if len(me.patterns) > 0 || len(me.prefixes) > 0 {
	// 	matched := false
	// 	for _, prefix := range me.prefixes {
	// 		if strings.HasPrefix(pathStr, prefix) {
	// 			matched = true
	// 			break
	// 		}
	// 	}
	// 	if !matched {
	// 		next(ctx, resp, req)
	// 		return
	// 	}
	// }

	me.handleCorsReq(ctx, resp, req, next)
}

func (me *CorzFilter) handleCorsReq(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	if req.Method == http.MethodOptions {
		h := make(http.Header)
		me.sendHeaders(h, req)
		resp(http.StatusNoContent, h)
		return
	}

	nResp := func(status int, headers http.Header) io.Writer {
		if headers == nil {
			headers = make(http.Header)
		}
		me.sendHeaders(headers, req)
		return resp(status, headers)
	}
	next(ctx, nResp, req)
}

func (me *CorzFilter) sendHeaders(ret http.Header, req *httpzrequest.Req) {
	origin := ""
	if req != nil && req.Headers != nil {
		origin = req.Headers.Get("Origin")
	}

	if origin != "" {
		ret.Set("Access-Control-Allow-Origin", origin)
		ret.Set("Access-Control-Allow-Credentials", "true")
		ret.Add("Vary", "Origin")
	} else {
		ret.Set("Access-Control-Allow-Origin", "*")
	}

	reqHeaders := ""
	if req != nil && req.Headers != nil {
		reqHeaders = req.Headers.Get("Access-Control-Request-Headers")
	}
	if reqHeaders != "" {
		ret.Set("Access-Control-Allow-Headers", reqHeaders)
	} else if v := me.Headers.Get("Access-Control-Allow-Headers"); v != "" {
		ret.Set("Access-Control-Allow-Headers", v)
	}

	if v := me.Headers.Get("Access-Control-Expose-Headers"); v != "" {
		ret.Set("Access-Control-Expose-Headers", v)
	}

	if len(me.Methods) > 0 {
		ret.Set("Access-Control-Allow-Methods", strings.Join(me.Methods, ", "))
	}

	ret.Set("Access-Control-Max-Age", "86400")

	if req != nil && req.Headers != nil && req.Headers.Get("Access-Control-Request-Private-Network") == "true" {
		ret.Set("Access-Control-Allow-Private-Network", "true")
	}
}

func (me *CorzFilter) fix() {
	if me.Headers == nil {
		me.mu.Lock()
		defer me.mu.Unlock()
		if me.Headers == nil {
			me.Headers = make(http.Header, 3)
			me.Headers.Set("Access-Control-Allow-Origin", "*")
			me.Headers.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Cache-Control, If-None-Match, Access-Control-Request-Private-Network")
			me.Headers.Set("Access-Control-Expose-Headers", "Etag, Md5Hash, Content-Length, Content-Size, Cache-Control")
		}
		if me.Methods == nil {
			me.Methods = []string{"GET", "POST", "OPTIONS", "PUT", "DELETE", "PATCH", "HEAD"}
		}
		if me.Methods != nil && me.Headers.Get("Access-Control-Allow-Methods") == "" {
			me.Headers.Set("Access-Control-Allow-Methods", strings.Join(me.Methods, ", "))
		}
	}
}
