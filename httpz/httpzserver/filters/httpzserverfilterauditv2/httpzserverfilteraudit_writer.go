package httpzserverfilterauditv2

import (
	"context"
	"io"
)

func AuditResponseBody(ctx context.Context, w io.Writer) io.Writer {
	return w
}
