// Package httpzserverfilterauthbasic provides HTTP Basic Authentication middleware for httpzserver,
// validating credentials using a callback function configured in Options.
package httpzserverfilterauthbasic

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

// Options configures the basic auth filter.
type Options struct {
	Verify func(ctx context.Context, user string, pass string) (bool, error)
	Realm  string
}

func (o *Options) fix() {
	if o.Realm == "" {
		o.Realm = "Restricted"
	}
}

// New creates an httpzserver.Filter that enforces HTTP Basic Authentication using the provided Options.
func New(ctx context.Context, opts Options) httpzserver.Filter {
	opts.fix()

	return func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
		user, pass, ok := parseBasicAuth(req)
		if !ok {
			unauthorized(resp, opts.Realm)
			return
		}
		if opts.Verify == nil {
			unauthorized(resp, opts.Realm)
			return
		}
		valid, err := opts.Verify(ctx, user, pass)
		if err != nil {
			panic(fmt.Errorf("error verifying credentials: %w", err))
		}
		if !valid {
			unauthorized(resp, opts.Realm)
			return
		}
		next(ctx, resp, req)
	}
}

func parseBasicAuth(req *httpzrequest.Req) (username, password string, ok bool) {
	if req == nil || req.Headers == nil {
		return "", "", false
	}
	auth := req.Headers.Get("Authorization")
	if auth == "" {
		return "", "", false
	}
	const prefix = "Basic "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return "", "", false
	}
	payload := strings.TrimSpace(auth[len(prefix):])
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", "", false
	}
	username, password, ok = strings.Cut(string(decoded), ":")
	if !ok {
		return "", "", false
	}
	return username, password, true
}

func unauthorized(resp httpzserver.Resp, realm string) {
	h := make(http.Header)
	h.Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w := resp(http.StatusUnauthorized, h)
	_, err := w.Write([]byte("Unauthorized"))
	errorz.Check(err)
}
