package schemazsamplefraction

import (
	"testing"

	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
)

func TestUnitBasic(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{})
	defer s.Close()
	s.Listen()
	s.Start()

}
