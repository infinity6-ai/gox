package httpzserverv2_test

import (
	"testing"

	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
)

func TestUnitListen(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()
	s.Listen()
	s.Start()

	// r, err := http.Get(s.Base())
	// errorz.Check(err)
	// defer r.Body.Close()
	// require.Equal(t, http.StatusNoContent, r.StatusCode)
}
