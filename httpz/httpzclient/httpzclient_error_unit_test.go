package httpzclient_test

import (
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

func TestUnitClientErrorHandling(t *testing.T) {
	ctx := t.Context()

	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()
	s.Listen()
	s.Start()
	s.Close()
	s.Close()

	t.Run("Request execution fails", func(t *testing.T) {
		// Using a non-routable IP address to simulate a network error
		client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: s.Base()})
		req := httpzrequest.NewReq("GET", "/")
		_, err := client.Do(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to execute request")
	})
}
