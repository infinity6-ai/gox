package httpzserver_test

import (
	"testing"

	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
)

func TestUnitListen(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()
	s.Listen()
}
