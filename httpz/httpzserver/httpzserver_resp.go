package httpzserver

import (
	"io"
	"net/http"
)

// Resp is a callback for writing the HTTP status code and response headers.
// It returns an io.Writer for streaming the response body back to the client.
type Resp func(status int, headers http.Header) io.Writer
