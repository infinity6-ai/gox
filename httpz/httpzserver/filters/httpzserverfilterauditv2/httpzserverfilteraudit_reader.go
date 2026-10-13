package httpzserverfilterauditv2

import (
	"context"
	"io"
)

func AuditRequestBody(ctx context.Context, r io.Reader) io.Reader {
	return r
}
