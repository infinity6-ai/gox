package schemazsamplefraction_test

import (
	"testing"

	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/infinity6-ai/gox/schemaz/schemazsamplefraction"
)

func TestUnitBasic(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{})
	defer s.Close()
	s.Listen()
	s.Start()

	schemazsamplefraction.Handlers(s)
}
