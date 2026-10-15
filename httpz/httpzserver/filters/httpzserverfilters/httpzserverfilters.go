package httpzserverfilters

import (
	"context"

	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilteraudit"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfiltererror"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfiltergzip"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilterlogz"
)

func BasicFilters(app string, version string) []httpzserver.Filter {
	filters := []httpzserver.Filter{}
	if app != "" || version != "" {
		filters = append(filters, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
			if app != "" {
				ctx = logz.SetAttr(ctx, logz.CtxKeyAppName, app)
			}
			if version != "" {
				ctx = logz.SetAttr(ctx, logz.CtxKeyAppVersion, version)
			}
			next(ctx, resp, req)
		})
	}
	filters = append(filters,
		httpzserverfilterlogz.Filter,
		httpzserverfilteraudit.Filter,
		httpzserverfiltergzip.GzipFilter,
		httpzserverfiltererror.Filter,
	)
	return filters
}

func Config(s *httpzserver.Server, app string, version string) {
	if app != "" {
		s.Context = logz.SetAttr(s.Context, logz.CtxKeyAppName, app)
	}
	if version != "" {
		s.Context = logz.SetAttr(s.Context, logz.CtxKeyAppVersion, version)
	}
	filters := BasicFilters(app, version)
	for _, f := range filters {
		s.AddFilter(f)
	}
}
