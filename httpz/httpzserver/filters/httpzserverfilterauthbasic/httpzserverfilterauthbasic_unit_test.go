package httpzserverfilterauthbasic_test

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilterauthbasic"
	"github.com/stretchr/testify/require"
)

func TestUnitFilterAuthBasic(t *testing.T) {
	validUser := "admin"
	validPass := "secret"

	defaultVerify := func(ctx context.Context, username, password string) (bool, error) {
		return username == validUser && password == validPass, nil
	}

	type testScenario struct {
		name                 string
		authHeader           string
		realm                string
		customVerify         func(ctx context.Context, user string, pass string) (bool, error)
		useNilVerify         bool
		expectedStatus       int
		expectedBody         string
		expectedAuthHeader   string
		handlerShouldExecute bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		verifyFn := defaultVerify
		if s.useNilVerify {
			verifyFn = nil
		} else if s.customVerify != nil {
			verifyFn = s.customVerify
		}

		filter := httpzserverfilterauthbasic.New(ctx, httpzserverfilterauthbasic.Options{
			Verify: verifyFn,
			Realm:  s.realm,
		})
		srv.AddFilter(filter)

		handlerExecuted := false
		srv.AddHandler(http.MethodGet, "/protected", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			handlerExecuted = true
			w := resp(http.StatusOK, nil)
			_, _ = w.Write([]byte("access granted"))
		})

		srv.Start()

		req, err := http.NewRequest(http.MethodGet, srv.Base().MustJoinPathString("/protected").String(), nil)
		require.NoError(t, err)

		if s.authHeader != "" {
			req.Header.Set("Authorization", s.authHeader)
		}

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, s.expectedStatus, resp.StatusCode)
		require.Equal(t, s.handlerShouldExecute, handlerExecuted)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, s.expectedBody, string(bodyBytes))

		if s.expectedAuthHeader != "" {
			require.Equal(t, s.expectedAuthHeader, resp.Header.Get("WWW-Authenticate"))
		}
	}

	encodeBasic := func(user, pass string) string {
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
	}

	t.Run("Valid credentials grants access", func(t *testing.T) {
		check(t, testScenario{
			authHeader:           encodeBasic(validUser, validPass),
			expectedStatus:       http.StatusOK,
			expectedBody:         "access granted",
			handlerShouldExecute: true,
		})
	})

	t.Run("Missing authorization header returns 401 with default realm", func(t *testing.T) {
		check(t, testScenario{
			authHeader:           "",
			expectedStatus:       http.StatusUnauthorized,
			expectedBody:         "Unauthorized",
			expectedAuthHeader:   `Basic realm="Restricted"`,
			handlerShouldExecute: false,
		})
	})

	t.Run("Missing authorization header returns 401 with custom realm", func(t *testing.T) {
		check(t, testScenario{
			authHeader:           "",
			realm:                "MySecureApp",
			expectedStatus:       http.StatusUnauthorized,
			expectedBody:         "Unauthorized",
			expectedAuthHeader:   `Basic realm="MySecureApp"`,
			handlerShouldExecute: false,
		})
	})

	t.Run("Wrong password returns 401", func(t *testing.T) {
		check(t, testScenario{
			authHeader:           encodeBasic(validUser, "wrongpassword"),
			expectedStatus:       http.StatusUnauthorized,
			expectedBody:         "Unauthorized",
			expectedAuthHeader:   `Basic realm="Restricted"`,
			handlerShouldExecute: false,
		})
	})

	t.Run("Wrong username returns 401", func(t *testing.T) {
		check(t, testScenario{
			authHeader:           encodeBasic("stranger", validPass),
			expectedStatus:       http.StatusUnauthorized,
			expectedBody:         "Unauthorized",
			expectedAuthHeader:   `Basic realm="Restricted"`,
			handlerShouldExecute: false,
		})
	})

	t.Run("Non-Basic scheme returns 401", func(t *testing.T) {
		check(t, testScenario{
			authHeader:           "Bearer some-jwt-token",
			expectedStatus:       http.StatusUnauthorized,
			expectedBody:         "Unauthorized",
			expectedAuthHeader:   `Basic realm="Restricted"`,
			handlerShouldExecute: false,
		})
	})

	t.Run("Malformed base64 returns 401", func(t *testing.T) {
		check(t, testScenario{
			authHeader:           "Basic !!!notvalidbase64!!!",
			expectedStatus:       http.StatusUnauthorized,
			expectedBody:         "Unauthorized",
			expectedAuthHeader:   `Basic realm="Restricted"`,
			handlerShouldExecute: false,
		})
	})

	t.Run("Credentials without colon separator returns 401", func(t *testing.T) {
		check(t, testScenario{
			authHeader:           "Basic " + base64.StdEncoding.EncodeToString([]byte("nocolonhere")),
			expectedStatus:       http.StatusUnauthorized,
			expectedBody:         "Unauthorized",
			expectedAuthHeader:   `Basic realm="Restricted"`,
			handlerShouldExecute: false,
		})
	})

	t.Run("Case-insensitive basic prefix supported", func(t *testing.T) {
		check(t, testScenario{
			authHeader:           "basic " + base64.StdEncoding.EncodeToString([]byte(validUser+":"+validPass)),
			expectedStatus:       http.StatusOK,
			expectedBody:         "access granted",
			handlerShouldExecute: true,
		})
	})

	t.Run("Password with colons handled correctly", func(t *testing.T) {
		check(t, testScenario{
			authHeader: encodeBasic("specialuser", "pass:with:multiple:colons"),
			customVerify: func(ctx context.Context, u, p string) (bool, error) {
				return u == "specialuser" && p == "pass:with:multiple:colons", nil
			},
			expectedStatus:       http.StatusOK,
			expectedBody:         "access granted",
			handlerShouldExecute: true,
		})
	})

	t.Run("Nil verify denies all requests", func(t *testing.T) {
		check(t, testScenario{
			authHeader:           encodeBasic(validUser, validPass),
			useNilVerify:         true,
			expectedStatus:       http.StatusUnauthorized,
			expectedBody:         "Unauthorized",
			expectedAuthHeader:   `Basic realm="Restricted"`,
			handlerShouldExecute: false,
		})
	})
}

func TestUnitFilterAuthBasic_VerifyError(t *testing.T) {
	ctx := t.Context()
	srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer srv.Close()
	srv.Listen()

	expectedErr := errors.New("db lookup failed")
	filter := httpzserverfilterauthbasic.New(ctx, httpzserverfilterauthbasic.Options{
		Verify: func(ctx context.Context, user string, pass string) (bool, error) {
			return false, expectedErr
		},
	})
	srv.AddFilter(filter)

	srv.AddHandler(http.MethodGet, "/err", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusOK, nil)
		_, _ = w.Write([]byte("ok"))
	})

	srv.Start()

	req, err := http.NewRequest(http.MethodGet, srv.Base().MustJoinPathString("/err").String(), nil)
	require.NoError(t, err)
	req.SetBasicAuth("admin", "secret")

	// Filter panics on Verify error, so the HTTP connection will be closed / fail without response
	client := &http.Client{}
	_, err = client.Do(req)
	require.Error(t, err)
}

func TestUnitFilterAuthBasic_NilRequest(t *testing.T) {
	filter := httpzserverfilterauthbasic.New(context.Background(), httpzserverfilterauthbasic.Options{
		Verify: func(ctx context.Context, u, p string) (bool, error) {
			return true, nil
		},
	})

	t.Run("Nil req returns 401", func(t *testing.T) {
		called := false
		next := func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req) {
			called = true
		}
		var status int
		resp := func(s int, h http.Header) io.Writer {
			status = s
			return io.Discard
		}
		filter(context.Background(), resp, nil, next)
		require.False(t, called)
		require.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("Req with nil Headers returns 401", func(t *testing.T) {
		called := false
		next := func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req) {
			called = true
		}
		var status int
		resp := func(s int, h http.Header) io.Writer {
			status = s
			return io.Discard
		}
		filter(context.Background(), resp, &httpzrequest.Req{Headers: nil}, next)
		require.False(t, called)
		require.Equal(t, http.StatusUnauthorized, status)
	})
}
