package httpzserverfilters

import (
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilteraudit"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfiltererror"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfiltergzip"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilterlogz"
)

func BasicFilters(app string, version string) []httpzserver.Filter {
	return []httpzserver.Filter{
		httpzserverfilterlogz.Filter,
		httpzserverfilteraudit.Filter,
		httpzserverfiltergzip.GzipFilter,
		httpzserverfiltererror.Filter,
	}
}

func Config(s *httpzserver.Server, app string, version string) {
	filters := BasicFilters(app, version)
	for _, f := range filters {
		s.AddFilter(f)
	}
}
