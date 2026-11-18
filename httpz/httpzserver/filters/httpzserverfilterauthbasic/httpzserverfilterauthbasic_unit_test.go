package httpzserverfilterauthbasic_test

import (
	"context"
	"encoding/base64"
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

	validator := func(ctx context.Context, username, password string) bool {
		return username == validUser && password == validPass
	}

	type testScenario struct {
		name                 string
		authHeader           string
		realm                string
		customCallback       httpzserverfilterauthbasic.Callback
		useNilCallback       bool
		useNew               bool
		expectedStatus       int
		expectedBody         string
		expectedAuthHeader   string
		expectedCtxUser      string
		handlerShouldExecute bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		var cb httpzserverfilterauthbasic.Callback
		if s.useNilCallback {
			cb = nil
		} else if s.customCallback != nil {
			cb = s.customCallback
		} else {
			cb = validator
		}

		var opts []httpzserverfilterauthbasic.Options
		if s.realm != "" {
			opts = append(opts, httpzserverfilterauthbasic.Options{Realm: s.realm})
		}

		var filter httpzserver.Filter
		if s.useNew {
			filter = httpzserverfilterauthbasic.New(cb, opts...)
		} else {
			filter = httpzserverfilterauthbasic.Filter(cb, opts...)
		}
		srv.AddFilter(filter)

		handlerExecuted := false
		var capturedUser string
		srv.AddHandler(http.MethodGet, "/protected", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			handlerExecuted = true
			capturedUser = httpzserverfilterauthbasic.User(ctx)
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

		if s.handlerShouldExecute {
			require.Equal(t, s.expectedCtxUser, capturedUser)
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
			expectedCtxUser:      validUser,
			handlerShouldExecute: true,
		})
	})

	t.Run("Valid credentials using New constructor", func(t *testing.T) {
		check(t, testScenario{
			authHeader:           encodeBasic(validUser, validPass),
			useNew:               true,
			expectedStatus:       http.StatusOK,
			expectedBody:         "access granted",
			expectedCtxUser:      validUser,
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
			expectedCtxUser:      validUser,
			handlerShouldExecute: true,
		})
	})

	t.Run("Password with colons handled correctly", func(t *testing.T) {
		check(t, testScenario{
			authHeader: encodeBasic("specialuser", "pass:with:multiple:colons"),
			customCallback: func(ctx context.Context, u, p string) bool {
				return u == "specialuser" && p == "pass:with:multiple:colons"
			},
			expectedStatus:       http.StatusOK,
			expectedBody:         "access granted",
			expectedCtxUser:      "specialuser",
			handlerShouldExecute: true,
		})
	})

	t.Run("Nil callback denies all requests", func(t *testing.T) {
		check(t, testScenario{
			authHeader:           encodeBasic(validUser, validPass),
			useNilCallback:       true,
			expectedStatus:       http.StatusUnauthorized,
			expectedBody:         "Unauthorized",
			expectedAuthHeader:   `Basic realm="Restricted"`,
			handlerShouldExecute: false,
		})
	})
}

func TestUnitFilterAuthBasic_Simple(t *testing.T) {
	t.Run("Simple callback grants access", func(t *testing.T) {
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		srv.AddFilter(httpzserverfilterauthbasic.Simple(func(user, pass string) bool {
			return user == "alice" && pass == "secret"
		}))

		srv.AddHandler(http.MethodGet, "/simple", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			w := resp(http.StatusOK, nil)
			_, _ = w.Write([]byte("hello " + httpzserverfilterauthbasic.UserFromContext(ctx)))
		})

		srv.Start()

		req, err := http.NewRequest(http.MethodGet, srv.Base().MustJoinPathString("/simple").String(), nil)
		require.NoError(t, err)
		req.SetBasicAuth("alice", "secret")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "hello alice", string(body))
	})

	t.Run("NewSimple constructor and rejection", func(t *testing.T) {
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		srv.AddFilter(httpzserverfilterauthbasic.NewSimple(func(user, pass string) bool {
			return user == "alice" && pass == "secret"
		}))

		srv.AddHandler(http.MethodGet, "/simple", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			w := resp(http.StatusOK, nil)
			_, _ = w.Write([]byte("ok"))
		})

		srv.Start()

		req, err := http.NewRequest(http.MethodGet, srv.Base().MustJoinPathString("/simple").String(), nil)
		require.NoError(t, err)
		req.SetBasicAuth("alice", "wrong")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Simple with nil callback denies", func(t *testing.T) {
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		srv.AddFilter(httpzserverfilterauthbasic.Simple(nil))

		srv.AddHandler(http.MethodGet, "/simple", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			w := resp(http.StatusOK, nil)
			_, _ = w.Write([]byte("ok"))
		})

		srv.Start()

		req, err := http.NewRequest(http.MethodGet, srv.Base().MustJoinPathString("/simple").String(), nil)
		require.NoError(t, err)
		req.SetBasicAuth("alice", "secret")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestUnitFilterAuthBasic_UserContext(t *testing.T) {
	t.Run("Nil context returns empty user", func(t *testing.T) {
		require.Empty(t, httpzserverfilterauthbasic.User(nil))
		require.Empty(t, httpzserverfilterauthbasic.UserFromContext(nil))
	})

	t.Run("Context without user returns empty user", func(t *testing.T) {
		ctx := context.Background()
		require.Empty(t, httpzserverfilterauthbasic.User(ctx))
	})

	t.Run("ContextWithUser sets user readable by User and UserFromContext", func(t *testing.T) {
		ctx := context.Background()
		ctx = httpzserverfilterauthbasic.ContextWithUser(ctx, "carol")
		require.Equal(t, "carol", httpzserverfilterauthbasic.User(ctx))
		require.Equal(t, "carol", httpzserverfilterauthbasic.UserFromContext(ctx))
	})
}

func TestUnitFilterAuthBasic_NilRequest(t *testing.T) {
	filter := httpzserverfilterauthbasic.Filter(func(ctx context.Context, u, p string) bool {
		return true
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
