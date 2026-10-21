package httpzserver

import (
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

// Resp is a callback for writing the HTTP status code and response headers.
// It returns an io.Writer for streaming the response body back to the client.
type Resp func(status int, headers http.Header) io.Writer

// Redirect sends an HTTP 307 Temporary Redirect response to the specified URL.
// If headers is nil, a new header map is created.
func (r Resp) Redirect(headers http.Header, u string) {
	if headers == nil {
		headers = make(http.Header)
	}
	headers.Set("Location", u)
	r(http.StatusTemporaryRedirect, headers)
}

// RedirectJS sends an HTTP 200 OK HTML response that redirects the browser to
// the specified URL using JavaScript on body load.
// If headers is nil, a new header map is created.
func (r Resp) RedirectJS(headers http.Header, u string) {
	if headers == nil {
		headers = make(http.Header)
	}
	headers.Set("Content-Type", "text/html; charset=UTF-8")
	w := r(http.StatusOK, headers)
	_, err := w.Write([]byte(`<html><body onload="window.location.href='` + u + `'"></body></html>`))
	errorz.Check(err)
}
