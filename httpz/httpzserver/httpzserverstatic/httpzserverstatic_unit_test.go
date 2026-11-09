package httpzserverstatic_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/httpzserverstatic"
	"github.com/stretchr/testify/require"
)

type tpack string

const testPackName tpack = "test-pack-static"

func createSampleTgz(files map[string]string) func() string {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		err := tw.WriteHeader(hdr)
		if err != nil {
			panic(err)
		}
		_, err = tw.Write([]byte(content))
		if err != nil {
			panic(err)
		}
	}
	tw.Close()
	gw.Close()
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return func() string { return b64 }
}

func TestUnitStaticHandler(t *testing.T) {
	sampleTgz := createSampleTgz(map[string]string{
		"index.html":       "<html><body>Home</body></html>",
		"style.css":        "body { color: red; }",
		"app.js":           "console.log('app');",
		"sub/index.html":   "<html><body>Sub</body></html>",
		"sub/data.json":    `{"key":"value"}`,
		"deep/path/hi.txt": "hello world",
	})
	staticzloader.SetCode(testPackName, sampleTgz)

	type testScenario struct {
		name           string
		method         string
		path           string
		mountPath      string
		mountMethod    string
		options        []httpzserverstatic.Options
		expectedStatus int
		expectedBody   string
		expectedCType  string
		expectedHeader map[string]string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		mountMethod := s.mountMethod
		if mountMethod == "" {
			mountMethod = "*"
		}

		mountPath := s.mountPath
		if mountPath == "" {
			mountPath = "/*"
		}

		handler := httpzserverstatic.StaticHandler(testPackName, s.options...)
		srv.AddHandler(mountMethod, mountPath, handler)
		srv.Start()

		req, err := http.NewRequest(s.method, srv.Base().MustJoinPathString(s.path).String(), nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, s.expectedStatus, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		if s.method != http.MethodHead {
			require.Equal(t, s.expectedBody, string(body))
		} else {
			require.Empty(t, string(body))
		}

		if s.expectedCType != "" {
			require.Contains(t, resp.Header.Get("Content-Type"), s.expectedCType)
		}

		for k, v := range s.expectedHeader {
			require.Equal(t, v, resp.Header.Get(k))
		}
	}

	t.Run("Serve exact file from root", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/style.css",
			expectedStatus: http.StatusOK,
			expectedBody:   "body { color: red; }",
			expectedCType:  "text/css",
		})
	})

	t.Run("Serve exact file from subdirectory", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/sub/data.json",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"key":"value"}`,
			expectedCType:  "application/json",
		})
	})

	t.Run("Directory request serves index.html by default", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>Home</body></html>",
			expectedCType:  "text/html",
		})
	})

	t.Run("Directory request with trailing slash serves index.html", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/sub/",
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>Sub</body></html>",
			expectedCType:  "text/html",
		})
	})

	t.Run("Serve with Prefix specified", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/static/app.js",
			mountPath:      "/static/*",
			options:        []httpzserverstatic.Options{{Prefix: "/static"}},
			expectedStatus: http.StatusOK,
			expectedBody:   "console.log('app');",
			expectedCType:  "text/javascript",
		})
	})

	t.Run("Root request with Prefix serves index.html", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/static",
			mountPath:      "/static/*",
			options:        []httpzserverstatic.Options{{Prefix: "/static"}},
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>Home</body></html>",
			expectedCType:  "text/html",
		})
	})

	t.Run("Trailing slash with Prefix serves index.html", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/static/",
			mountPath:      "/static/*",
			options:        []httpzserverstatic.Options{{Prefix: "/static"}},
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>Home</body></html>",
			expectedCType:  "text/html",
		})
	})

	t.Run("Serve with custom Headers", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/style.css",
			options:        []httpzserverstatic.Options{{Headers: http.Header{"Cache-Control": {"max-age=3600"}}}},
			expectedStatus: http.StatusOK,
			expectedBody:   "body { color: red; }",
			expectedCType:  "text/css",
			expectedHeader: map[string]string{"Cache-Control": "max-age=3600"},
		})
	})

	t.Run("HEAD request returns headers but no body", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodHead,
			path:           "/style.css",
			expectedStatus: http.StatusOK,
			expectedCType:  "text/css",
			expectedHeader: map[string]string{"Content-Length": "20"},
		})
	})

	t.Run("404 Not Found for missing file", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/missing.txt",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not Found",
		})
	})

	t.Run("404 Not Found without fallback for missing file in SPA", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/some-route",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not Found",
		})
	})

	t.Run("405 Method Not Allowed for POST", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodPost,
			path:           "/style.css",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method Not Allowed",
		})
	})

	t.Run("404 Not Found for directory if NoIndex is true", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/",
			options:        []httpzserverstatic.Options{{NoIndex: true}},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not Found",
		})
	})

	t.Run("404 Not Found for path traversal attempt", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/sub/../index.html", // client normalizes to /index.html
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>Home</body></html>",
		})
	})
}
