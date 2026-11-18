package httpzsite_test

import (
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzsite"
	"github.com/stretchr/testify/require"
)

func TestUnitSiteBasicUsers(t *testing.T) {
	ctx := t.Context()

	t.Run("Site with BasicUsers enforces authentication", func(t *testing.T) {
		site := &httpzsite.Site{
			Context:      ctx,
			LocalAddress: "localhost:0",
			Pack:         testPackName,
			BasicUsers: map[string]string{
				"admin": "secret123",
				"guest": "welcome",
			},
		}

		site.Listen()
		site.AddHandlers()
		site.Start()
		defer site.Close()

		base := site.Base().String()

		// Request without credentials fails with 401
		resp, err := http.Get(base)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		// Request with invalid password fails with 401
		req, err := http.NewRequest(http.MethodGet, base, nil)
		require.NoError(t, err)
		req.SetBasicAuth("admin", "wrong")
		resp2, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp2.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp2.StatusCode)

		// Request with unknown user fails with 401
		req, err = http.NewRequest(http.MethodGet, base, nil)
		require.NoError(t, err)
		req.SetBasicAuth("unknown", "secret123")
		resp3, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp3.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp3.StatusCode)

		// Request with valid credentials for admin succeeds with 200
		req, err = http.NewRequest(http.MethodGet, base, nil)
		require.NoError(t, err)
		req.SetBasicAuth("admin", "secret123")
		resp4, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp4.Body.Close()
		require.Equal(t, http.StatusOK, resp4.StatusCode)

		// Request with valid credentials for guest succeeds with 200
		req, err = http.NewRequest(http.MethodGet, base+"style.css", nil)
		require.NoError(t, err)
		req.SetBasicAuth("guest", "welcome")
		resp5, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp5.Body.Close()
		require.Equal(t, http.StatusOK, resp5.StatusCode)
	})

	t.Run("Site with nil BasicUsers allows anonymous access", func(t *testing.T) {
		site := &httpzsite.Site{
			Context:      ctx,
			LocalAddress: "localhost:0",
			Pack:         testPackName,
			BasicUsers:   nil,
		}

		site.Listen()
		site.AddHandlers()
		site.Start()
		defer site.Close()

		base := site.Base().String()

		resp, err := http.Get(base)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
