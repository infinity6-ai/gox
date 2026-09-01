package httpzserver_test

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/stretchr/testify/require"
)

func TestUnitServerLifecycle(t *testing.T) {
	ctx := t.Context()

	t.Run("Listen and get address", func(t *testing.T) {
		s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer s.Close()

		require.Nil(t, s.Addr()) // Address should be nil before Listen
		s.Listen()
		require.NotNil(t, s.Addr())
		require.True(t, strings.HasPrefix(s.Addr().String(), "127.0.0.1:") || strings.HasPrefix(s.Addr().String(), "[::1]:"), "address should be local")

		addr, ok := s.Addr().(*net.TCPAddr)
		require.True(t, ok)
		require.NotZero(t, addr.Port, "port should not be zero after listen")

		base := s.Base()
		require.Equal(t, fmt.Sprintf("http://%s/", s.Addr()), base.String())
	})

	t.Run("Double listen panics", func(t *testing.T) {
		s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer s.Close()
		s.Listen()
		require.Panics(t, func() {
			s.Listen()
		}, "calling Listen twice should panic")
	})

	t.Run("Start without listen panics", func(t *testing.T) {
		s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer s.Close()
		require.Panics(t, func() {
			s.Start()
		}, "calling Start before Listen should panic")
	})

	t.Run("Serve without listen panics", func(t *testing.T) {
		s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer s.Close()
		require.Panics(t, func() {
			s.Serve()
		}, "calling Serve before Listen should panic")
	})
}
