package httpzhelper

import (
	"net/http"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
)

func FromHttpRequest(input *http.Request, output *httpzrequest.Req) {
	output.Method = input.Method
	output.Path = pathz.MustParse(input.URL.Path)
	checker.True(output.Path.IsAbsolute(), "must be absolute: %s", output.Path)
	output.Query = input.URL.Query()
	output.Headers = input.Header
	output.Body = input.Body
}
