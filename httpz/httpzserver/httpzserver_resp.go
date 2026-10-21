package httpzserver

import (
	"io"
	"net/http"
)

// Resp is a callback for writing the HTTP status code and response headers.
// It returns an io.Writer for streaming the response body back to the client.
type Resp func(status int, headers http.Header) io.Writer

func (r Resp) Redirect(headers http.Header, u string) {
	headers.Set("Location", u)
	r(http.StatusTemporaryRedirect, headers)
}

func (r Resp) RedirectJS(headers http.Header, u string) {
	headers.Set("Content-Type", "text/html; charset=UTF-8")
	w := r(http.StatusOK, headers)
	w.Write([]byte(`<html>body onload="` + u + `"></body></html>`))
}
