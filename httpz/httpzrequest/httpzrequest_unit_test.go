package httpzrequest_test

import (
	"io"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/stretchr/testify/require"
)

func TestUnitRequestFormat(t *testing.T) {
	type testScenario struct {
		name           string
		method         string
		pathPattern    string
		params         map[string]string
		expectedPath   string
		expectedErrMsg string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		req, err := httpzrequest.Format(s.method, s.pathPattern, s.params)
		if s.expectedErrMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErrMsg)
			return
		}

		require.NoError(t, err)
		require.Equal(t, s.method, req.Method)
		require.Equal(t, s.expectedPath, req.Path.String())
	}

	t.Run("Format with parameters", func(t *testing.T) {
		check(t, testScenario{
			name:         "Format with parameters",
			method:       "GET",
			pathPattern:  "/users/{id}/profile",
			params:       map[string]string{"id": "123"},
			expectedPath: "/users/123/profile",
		})
	})

	t.Run("Format without parameters", func(t *testing.T) {
		check(t, testScenario{
			name:         "Format without parameters",
			method:       "POST",
			pathPattern:  "/users/list",
			params:       nil,
			expectedPath: "/users/list",
		})
	})

	t.Run("Format with missing parameter error", func(t *testing.T) {
		check(t, testScenario{
			name:           "Format with missing parameter error",
			method:         "GET",
			pathPattern:    "/users/{id}",
			params:         map[string]string{"other": "val"},
			expectedErrMsg: "failed to format path",
		})
	})
}

func TestUnitRequestMustFormat(t *testing.T) {
	type testScenario struct {
		name         string
		method       string
		path         string
		params       map[string]string
		expectPanic  bool
		expectedPath string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.expectPanic {
			require.Panics(t, func() {
				httpzrequest.MustFormat(s.method, s.path, s.params)
			})
			return
		}

		req := httpzrequest.MustFormat(s.method, s.path, s.params)
		require.NotNil(t, req)
		require.Equal(t, s.expectedPath, req.Path.String())
	}

	t.Run("MustFormat succeeds on valid path and params", func(t *testing.T) {
		check(t, testScenario{
			name:         "MustFormat succeeds on valid path and params",
			method:       "GET",
			path:         "/items/{name}",
			params:       map[string]string{"name": "book"},
			expectPanic:  false,
			expectedPath: "/items/book",
		})
	})

	t.Run("MustFormat panics on missing required param", func(t *testing.T) {
		check(t, testScenario{
			name:        "MustFormat panics on missing required param",
			method:      "GET",
			path:        "/items/{name}",
			params:      map[string]string{"other": "val"},
			expectPanic: true,
		})
	})
}

func TestUnitRequestBuilders(t *testing.T) {
	t.Run("New initializes fields", func(t *testing.T) {
		req := httpzrequest.New("GET", "/test/path")
		require.Equal(t, "GET", req.Method)
		require.Equal(t, "/test/path", req.Path.String())
		require.NotNil(t, req.Query)
		require.NotNil(t, req.Headers)
		require.Nil(t, req.Body)
	})

	t.Run("FromUrl initializes fields", func(t *testing.T) {
		u := urlz.MustParse("https://example.com/api")
		req := httpzrequest.FromUrl("POST", u)
		require.Equal(t, "POST", req.Method)
		require.Equal(t, u, req.Url)
		require.NotNil(t, req.Query)
		require.NotNil(t, req.Headers)
	})

	t.Run("Query and header manipulation", func(t *testing.T) {
		req := httpzrequest.New("GET", "/").
			SetQuery("k", "v1").
			AddQuery("k", "v2").
			SetHeader("Content-Type", "application/json").
			AddHeader("X-Tag", "tag1").
			SetBody(strings.NewReader("hello"))

		require.Equal(t, []string{"v1", "v2"}, req.Query["k"])
		require.Equal(t, "application/json", req.Headers.Get("Content-Type"))
		require.Equal(t, "tag1", req.Headers.Get("X-Tag"))

		b, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.Equal(t, "hello", string(b))
	})
}

func TestUnitRequestResolveUrl(t *testing.T) {
	baseUrl := urlz.MustParse("https://api.example.com/v1")
	subUrl := urlz.MustParse("https://api.example.com/v1/users")
	otherUrl := urlz.MustParse("https://other.com/users")

	type testScenario struct {
		name           string
		path           *pathz.Path
		url            *urlz.Url
		base           *urlz.Url
		expectPanic    bool
		expectedResult string
		expectedErrMsg string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		req := &httpzrequest.Req{
			Path: s.path,
			Url:  s.url,
		}

		if s.expectPanic {
			require.Panics(t, func() {
				_, _ = req.ResolveUrl(s.base)
			})
			return
		}

		resolved, err := req.ResolveUrl(s.base)
		if s.expectedErrMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErrMsg)
			return
		}

		require.NoError(t, err)
		require.Equal(t, s.expectedResult, resolved.String())
	}

	t.Run("Neither Path nor Url panics", func(t *testing.T) {
		check(t, testScenario{
			name:        "Neither Path nor Url panics",
			base:        baseUrl,
			expectPanic: true,
		})
	})

	t.Run("Both Path and Url panics", func(t *testing.T) {
		check(t, testScenario{
			name:        "Both Path and Url panics",
			path:        pathz.MustParse("/path"),
			url:         subUrl,
			base:        baseUrl,
			expectPanic: true,
		})
	})

	t.Run("Path without base URL fails", func(t *testing.T) {
		check(t, testScenario{
			name:           "Path without base URL fails",
			path:           pathz.MustParse("/users"),
			base:           nil,
			expectedErrMsg: "base url not found",
		})
	})

	t.Run("Path with base URL resolves", func(t *testing.T) {
		check(t, testScenario{
			name:           "Path with base URL resolves",
			path:           pathz.MustParse("users"),
			base:           urlz.MustParse("https://api.example.com/v1/"),
			expectedResult: "https://api.example.com/v1/users",
		})
	})

	t.Run("Url matching base URL succeeds", func(t *testing.T) {
		check(t, testScenario{
			name:           "Url matching base URL succeeds",
			url:            subUrl,
			base:           baseUrl,
			expectedResult: "https://api.example.com/v1/users",
		})
	})

	t.Run("Url mismatching base URL fails", func(t *testing.T) {
		check(t, testScenario{
			name:           "Url mismatching base URL fails",
			url:            otherUrl,
			base:           baseUrl,
			expectedErrMsg: "base url mismatch",
		})
	})

	t.Run("Url with nil base URL succeeds", func(t *testing.T) {
		check(t, testScenario{
			name:           "Url with nil base URL succeeds",
			url:            otherUrl,
			base:           nil,
			expectedResult: "https://other.com/users",
		})
	})
}
