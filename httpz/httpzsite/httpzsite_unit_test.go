package httpzsite_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
	"github.com/infinity6-ai/gox/httpz/httpzsite"
	"github.com/stretchr/testify/require"
)

type tpack string

const testPackName tpack = "test-pack-site"

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

func init() {
	sampleTgz := createSampleTgz(map[string]string{
		"index.html":        "<html><body>Home</body></html>",
		"style.css":         "body { color: red; }",
		"sub/index.html":    "<html><body>Sub</body></html>",
		"sub/script.js":     "console.log('sub');",
		"deep/nested/a.txt": "deep-content",
	})
	staticzloader.SetCode(testPackName, sampleTgz)
}

func TestUnitSiteStaticFiles(t *testing.T) {
	ctx := t.Context()

	site := &httpzsite.Site{
		Context:      ctx,
		LocalAddress: "localhost:0",
		Pack:         testPackName,
	}
	site.Listen()
	site.AddHandlers()
	site.Start()
	defer site.Close()

	baseURL := site.Base().String()

	type testScenario struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		req, err := http.NewRequest(s.method, baseURL+s.path, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, s.expectedStatus, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		if s.expectedBody != "" {
			require.Equal(t, s.expectedBody, string(body))
		}
	}

	t.Run("Root path serves index.html", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "",
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>Home</body></html>",
		})
	})

	t.Run("Explicit index.html path", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "index.html",
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>Home</body></html>",
		})
	})

	t.Run("Static css asset", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "style.css",
			expectedStatus: http.StatusOK,
			expectedBody:   "body { color: red; }",
		})
	})

	t.Run("Subdirectory trailing slash serves sub index.html", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "sub/",
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>Sub</body></html>",
		})
	})

	t.Run("Subdirectory explicit index.html", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "sub/index.html",
			expectedStatus: http.StatusOK,
			expectedBody:   "<html><body>Sub</body></html>",
		})
	})

	t.Run("Subdirectory javascript asset", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "sub/script.js",
			expectedStatus: http.StatusOK,
			expectedBody:   "console.log('sub');",
		})
	})

	t.Run("Deeply nested file", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "deep/nested/a.txt",
			expectedStatus: http.StatusOK,
			expectedBody:   "deep-content",
		})
	})

	t.Run("Non-existent file returns 404", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodGet,
			path:           "nonexistent.html",
			expectedStatus: http.StatusNotFound,
		})
	})

	t.Run("Non-GET method returns 404", func(t *testing.T) {
		check(t, testScenario{
			method:         http.MethodPost,
			path:           "style.css",
			expectedStatus: http.StatusNotFound,
		})
	})
}

func TestUnitSiteLifecycle(t *testing.T) {
	t.Run("Unstarted site returns nil addr and base and close succeeds", func(t *testing.T) {
		site := &httpzsite.Site{
			Context:      t.Context(),
			LocalAddress: "localhost:0",
			Pack:         testPackName,
		}

		require.Nil(t, site.Addr())
		require.Nil(t, site.Base())
		require.NoError(t, site.Close())
	})

	t.Run("Listen populates Addr and Base before Start", func(t *testing.T) {
		site := &httpzsite.Site{
			Context:      t.Context(),
			LocalAddress: "localhost:0",
			Pack:         testPackName,
		}
		defer site.Close()

		require.Nil(t, site.Addr())
		require.Nil(t, site.Base())

		site.Listen()

		addr := site.Addr()
		require.NotNil(t, addr)
		tcpAddr, ok := addr.(*net.TCPAddr)
		require.True(t, ok)
		require.NotZero(t, tcpAddr.Port)

		base := site.Base()
		require.NotNil(t, base)
		require.NotEmpty(t, base.String())
	})

	t.Run("Double listen panics", func(t *testing.T) {
		site := &httpzsite.Site{
			Context:      t.Context(),
			LocalAddress: "localhost:0",
			Pack:         testPackName,
		}
		defer site.Close()

		site.Listen()
		require.Panics(t, func() {
			site.Listen()
		})
	})

	t.Run("Start without listen panics", func(t *testing.T) {
		site := &httpzsite.Site{
			Context:      t.Context(),
			LocalAddress: "localhost:0",
			Pack:         testPackName,
		}
		defer site.Close()

		require.Panics(t, func() {
			site.Start()
		})
	})

	t.Run("Serve without listen panics", func(t *testing.T) {
		site := &httpzsite.Site{
			Context:      t.Context(),
			LocalAddress: "localhost:0",
			Pack:         testPackName,
		}
		defer site.Close()

		require.Panics(t, func() {
			site.Serve()
		})
	})

	t.Run("Start and Close lifecycle", func(t *testing.T) {
		site := &httpzsite.Site{
			Context:      t.Context(),
			LocalAddress: "localhost:0",
			Pack:         testPackName,
		}

		site.Listen()
		site.AddHandlers()
		site.Start()
		defer site.Close()

		base := site.Base()
		require.NotNil(t, base)

		// Verify server answers requests
		resp, err := http.Get(base.String())
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Close server
		require.NoError(t, site.Close())

		// Subsequent close returns nil
		require.NoError(t, site.Close())

		// Verify requests now fail
		client := &http.Client{Timeout: 500 * time.Millisecond}
		_, err = client.Get(base.String())
		require.Error(t, err)
	})

	t.Run("Serve lifecycle and clean shutdown", func(t *testing.T) {
		site := &httpzsite.Site{
			Context:      t.Context(),
			LocalAddress: "localhost:0",
			Pack:         testPackName,
		}

		site.Listen()
		site.AddHandlers()
		defer site.Close()

		base := site.Base()
		require.NotNil(t, base)

		go site.Serve()

		resp, err := http.Get(base.String())
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.NoError(t, site.Close())
	})
}

func TestUnitSiteExtractTo(t *testing.T) {
	ctx := t.Context()

	t.Run("Extract all pack files to destination directory", func(t *testing.T) {
		site := &httpzsite.Site{
			Context: ctx,
			Pack:    testPackName,
		}

		tempDir := t.TempDir()
		dest := pathz.MustParse(tempDir)

		err := site.ExtractTo(ctx, dest)
		require.NoError(t, err)

		expectedFiles := map[string]string{
			"index.html":        "<html><body>Home</body></html>",
			"style.css":         "body { color: red; }",
			"sub/index.html":    "<html><body>Sub</body></html>",
			"sub/script.js":     "console.log('sub');",
			"deep/nested/a.txt": "deep-content",
		}

		for relPath, expectedContent := range expectedFiles {
			p := dest.MustJoin(pathz.MustParse(relPath))
			content, err := os.ReadFile(p.String())
			require.NoError(t, err, "failed to read extracted file: %s", p.String())
			require.Equal(t, expectedContent, string(content))
		}
	})

	t.Run("Extract invalid pack returns error", func(t *testing.T) {
		site := &httpzsite.Site{
			Context: ctx,
			Pack:    "nonexistent-pack",
		}

		tempDir := t.TempDir()
		dest := pathz.MustParse(tempDir)

		err := site.ExtractTo(ctx, dest)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to extract site pack")
	})
}

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
