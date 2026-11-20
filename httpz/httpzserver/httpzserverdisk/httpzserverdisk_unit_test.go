package httpzserverdisk_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/httpzserverdisk"
	"github.com/stretchr/testify/require"
)

func createTestDiskFiles(t *testing.T, files map[string]string) *pathz.Path {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "httpzserverdisk-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	for rel, content := range files {
		fullPath := filepath.Join(tempDir, rel)
		err := os.MkdirAll(filepath.Dir(fullPath), 0o755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(content), 0o644)
		require.NoError(t, err)
	}

	return pathz.MustParse(tempDir)
}

func TestUnitDiskHandler(t *testing.T) {
	baseDir := createTestDiskFiles(t, map[string]string{
		"index.html":       "<html><body>Home</body></html>",
		"style.css":        "body { color: red; }",
		"app.js":           "console.log('app');",
		"sub/index.html":   "<html><body>Sub</body></html>",
		"sub/data.json":    `{"key":"value"}`,
		"deep/path/hi.txt": "hello world",
	})

	type testScenario struct {
		name              string
		method            string
		path              string
		mountPath         string
		mountMethod       string
		options           httpzserverdisk.Options
		expectedStatus    int
		expectedBody      string
		expectedBodyMatch []string
		expectedCType     string
		expectedHeader    map[string]string
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

		opts := s.options
		if opts.Base == nil {
			opts.Base = baseDir
		}

		handler := httpzserverdisk.DiskHandler(opts)
		srv.AddHandler(mountMethod, mountPath, handler)
		srv.Start()

		baseURL, err := url.Parse(srv.Base().String())
		require.NoError(t, err)
		subURL, err := url.Parse(s.path)
		require.NoError(t, err)

		req, err := http.NewRequest(s.method, baseURL.ResolveReference(subURL).String(), nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, s.expectedStatus, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		if s.method != http.MethodHead {
			if s.expectedBody != "" {
				require.Equal(t, s.expectedBody, string(body))
			}
			for _, match := range s.expectedBodyMatch {
				require.Contains(t, string(body), match)
			}
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

	t.Run("Directory request without trailing slash serves index.html", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/sub",
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
			options:        httpzserverdisk.Options{Prefix: "/static"},
			expectedStatus: http.StatusOK,
			expectedBody:   "console.log('app');",
			expectedCType:  "text/javascript",
		})
	})

	t.Run("Prefix without leading slash is normalized", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/static/app.js",
			mountPath:      "/static/*",
			options:        httpzserverdisk.Options{Prefix: "static"},
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
			options:        httpzserverdisk.Options{Prefix: "/static"},
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
			options:        httpzserverdisk.Options{Prefix: "/static"},
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>Home</body></html>",
			expectedCType:  "text/html",
		})
	})

	t.Run("Serve with custom Headers", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/style.css",
			options:        httpzserverdisk.Options{Headers: http.Header{"Cache-Control": {"max-age=3600"}}},
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

	t.Run("405 Method Not Allowed for POST", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodPost,
			path:           "/style.css",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method Not Allowed",
		})
	})

	t.Run("404 Not Found for directory if NoIndex is true and Listing is false", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/",
			options:        httpzserverdisk.Options{NoIndex: true, Listing: false},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not Found",
		})
	})

	t.Run("404 Not Found for directory without index file when Listing is false", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/deep",
			options:        httpzserverdisk.Options{Listing: false},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not Found",
		})
	})

	t.Run("404 Not Found for directory trailing slash without index when Listing is false", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/deep/",
			options:        httpzserverdisk.Options{Listing: false},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not Found",
		})
	})

	t.Run("Directory listing rendered when NoIndex is true and Listing is true", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/",
			options:        httpzserverdisk.Options{NoIndex: true, Listing: true},
			expectedStatus: http.StatusOK,
			expectedCType:  "text/html",
			expectedBodyMatch: []string{
				"<title>Index of /</title>",
				"<h1>Index of /</h1>",
				`<a href="/style.css">style.css</a>`,
				`<a href="/app.js">app.js</a>`,
				`<a href="/sub/">sub/</a>`,
			},
		})
	})

	t.Run("Directory listing rendered for subdirectory without index (trailing slash)", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/deep/",
			options:        httpzserverdisk.Options{Listing: true},
			expectedStatus: http.StatusOK,
			expectedCType:  "text/html",
			expectedBodyMatch: []string{
				"<title>Index of /deep/</title>",
				"<h1>Index of /deep/</h1>",
				`<a href="/">../</a>`,
				`<a href="/deep/path/">path/</a>`,
			},
		})
	})

	t.Run("Directory listing rendered for subdirectory without index (no trailing slash)", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/deep",
			options:        httpzserverdisk.Options{Listing: true},
			expectedStatus: http.StatusOK,
			expectedCType:  "text/html",
			expectedBodyMatch: []string{
				"<title>Index of /deep/</title>",
				"<h1>Index of /deep/</h1>",
				`<a href="/">../</a>`,
				`<a href="/deep/path/">path/</a>`,
			},
		})
	})

	t.Run("Directory listing with Prefix preserves prefix in links", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/files/deep/",
			mountPath:      "/files/*",
			options:        httpzserverdisk.Options{Prefix: "/files", Listing: true},
			expectedStatus: http.StatusOK,
			expectedCType:  "text/html",
			expectedBodyMatch: []string{
				"<title>Index of /files/deep/</title>",
				"<h1>Index of /files/deep/</h1>",
				`<a href="/files/">../</a>`,
				`<a href="/files/deep/path/">path/</a>`,
			},
		})
	})

	t.Run("HEAD request on directory listing returns headers with empty body", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodHead,
			path:           "/deep/",
			options:        httpzserverdisk.Options{Listing: true},
			expectedStatus: http.StatusOK,
			expectedCType:  "text/html",
		})
	})

	t.Run("Directory listing as JSON flat (t=json)", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/?t=json",
			options:        httpzserverdisk.Options{Listing: true},
			expectedStatus: http.StatusOK,
			expectedCType:  "application/json",
			expectedBodyMatch: []string{
				`"path": "/"`,
				`"recursive": false`,
				`"name": "app.js"`,
				`"path": "/app.js"`,
				`"dir": false`,
				`"name": "sub"`,
				`"path": "/sub/"`,
				`"dir": true`,
			},
		})
	})

	t.Run("Directory listing as JSON recursive (t=json&r=true)", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/?t=json&r=true",
			options:        httpzserverdisk.Options{Listing: true},
			expectedStatus: http.StatusOK,
			expectedCType:  "application/json",
			expectedBodyMatch: []string{
				`"path": "/"`,
				`"recursive": true`,
				`"name": "app.js"`,
				`"name": "deep/path/hi.txt"`,
				`"path": "/deep/path/hi.txt"`,
				`"name": "sub/data.json"`,
				`"path": "/sub/data.json"`,
			},
		})
	})

	t.Run("Directory listing as JSON recursive with r=1", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/?t=json&r=1",
			options:        httpzserverdisk.Options{Listing: true},
			expectedStatus: http.StatusOK,
			expectedCType:  "application/json",
			expectedBodyMatch: []string{
				`"path": "/"`,
				`"recursive": true`,
				`"name": "deep/path/hi.txt"`,
			},
		})
	})

	t.Run("Directory listing as HTML recursive (r=true)", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/?r=true",
			options:        httpzserverdisk.Options{Listing: true},
			expectedStatus: http.StatusOK,
			expectedCType:  "text/html",
			expectedBodyMatch: []string{
				"<title>Index of / (recursive)</title>",
				"<h1>Index of / (recursive)</h1>",
				`<a href="/deep/path/hi.txt">deep/path/hi.txt</a>`,
				`<a href="/sub/data.json">sub/data.json</a>`,
			},
		})
	})

	t.Run("Explicit listing query bypasses index.html when Listing is true", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/?t=json",
			options:        httpzserverdisk.Options{Listing: true},
			expectedStatus: http.StatusOK,
			expectedCType:  "application/json",
			expectedBodyMatch: []string{
				`"name": "index.html"`,
			},
		})
	})

	t.Run("400 Bad Request for invalid t parameter", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/?t=xml",
			options:        httpzserverdisk.Options{Listing: true},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid format parameter 't', expected 'html' or 'json'",
		})
	})

	t.Run("400 Bad Request for invalid r parameter", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "/?r=notabool",
			options:        httpzserverdisk.Options{Listing: true},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid recursive parameter 'r', expected boolean",
		})
	})

	t.Run("HEAD request on JSON listing returns headers with empty body", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodHead,
			path:           "/?t=json",
			options:        httpzserverdisk.Options{Listing: true},
			expectedStatus: http.StatusOK,
			expectedCType:  "application/json",
		})
	})

	t.Run("Handler synonym works identical to DiskHandler", func(t *testing.T) {
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		srv.AddHandler("*", "/*", httpzserverdisk.Handler(httpzserverdisk.Options{Base: baseDir}))
		srv.Start()

		resp, err := http.Get(srv.Base().MustJoinPathString("/style.css").String())
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "body { color: red; }", string(body))
	})
}

func TestUnitOptionsValidation(t *testing.T) {
	t.Run("Panics when Base is nil", func(t *testing.T) {
		require.Panics(t, func() {
			httpzserverdisk.DiskHandler(httpzserverdisk.Options{
				Base: nil,
			})
		})
	})
}

func TestUnitListingJSONUnmarshal(t *testing.T) {
	baseDir := createTestDiskFiles(t, map[string]string{
		"a.txt":     "aaa",
		"sub/b.txt": "bbb",
	})

	ctx := t.Context()
	srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer srv.Close()
	srv.Listen()

	srv.AddHandler("*", "/*", httpzserverdisk.DiskHandler(httpzserverdisk.Options{
		Base:    baseDir,
		Listing: true,
	}))
	srv.Start()

	baseURL, err := url.Parse(srv.Base().String())
	require.NoError(t, err)
	subURL, err := url.Parse("/?t=json&r=true")
	require.NoError(t, err)

	resp, err := http.Get(baseURL.ResolveReference(subURL).String())
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listing httpzserverdisk.ListingResponse
	err = json.NewDecoder(resp.Body).Decode(&listing)
	require.NoError(t, err)

	require.Equal(t, "/", listing.Path)
	require.True(t, listing.Recursive)
	require.NotEmpty(t, listing.Entries)

	foundSubFile := false
	for _, entry := range listing.Entries {
		if entry.Name == "sub/b.txt" {
			foundSubFile = true
			require.False(t, entry.Dir)
			require.Equal(t, int64(3), entry.Size)
			require.Equal(t, "/sub/b.txt", entry.Path)
		}
	}
	require.True(t, foundSubFile, "expected to find sub/b.txt in recursive listing")
}
