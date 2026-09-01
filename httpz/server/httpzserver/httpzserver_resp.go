package httpzserver

import (
	"io"
	"net/http"
)

type Resp func(status int, headers http.Header) io.Writer
