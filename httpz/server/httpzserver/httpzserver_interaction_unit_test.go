package httpzserver_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/stretchr/testify/require"
)

// TestUnitFilterAndHandlerInteraction is a comprehensive integration-style unit test
// that verifies the complex interactions between filters, response wrapping, and the final handler.
func TestUnitFilterAndHandlerInteraction(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()
	s.Listen()

	// First filter: wraps the request body and modifies response headers.
	s.AddFilter(func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, next httpzserver.Handler) {
		preBody := strings.NewReader("reqpre-")
		originalBody := req.Body
		sufBody := strings.NewReader("-reqsuf")
		req.Body = io.MultiReader(preBody, originalBody, sufBody)

		nResp := func(status int, nHeaders http.Header) io.Writer {
			nHeaders.Set("a", "x1") // Set by filter 1
			nHeaders.Set("b", "x1") // Set by filter 1, should be overridden by filter 2
			return resp(status, nHeaders)
		}
		next(ctx, nResp, req)
	})

	// Second filter: uses WrapResponse to modify headers and inject content around the handler's output.
	s.AddFilter(func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, next httpzserver.Handler) {
		w := next.WrapResponse(ctx, resp, req, func(outStatus int, outHeaders http.Header) int {
			outHeaders.Set("b", "x2") // Overrides header from filter 1
			outHeaders.Set("c", "x2") // Set by filter 2
			return outStatus
		}, func(outWriter io.Writer) io.Writer {
			_, err := outWriter.Write([]byte("UUUU "))
			errorz.Check(err)
			return outWriter
		})
		_, err := w.Write([]byte(" ZZZZ"))
		errorz.Check(err)
	})

	// Handler for the specific route being tested.
	s.AddHandler("POST", "/bla/{p1}/b/{p2}/c/*", func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
		reqBody, err := io.ReadAll(req.Body)
		errorz.Check(err)

		// Create a sorted string representation of params for stable output.
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sortedParams := make([]string, len(keys))
		for i, k := range keys {
			sortedParams[i] = fmt.Sprintf("%s:%s", k, params[k])
		}
		paramsStr := fmt.Sprintf("map[%s]", strings.Join(sortedParams, " "))

		body := fmt.Appendf(nil, "r: %s - %s, req: %s", req.Path, paramsStr, string(reqBody))

		w := resp(http.StatusBadRequest, http.Header{
			"a": []string{"y1"}, // Set by handler, should be overridden by filter 1
			"d": []string{"y"},  // Set by handler
		})
		_, err = w.Write(body)
		errorz.Check(err)
	})

	s.Start()

	// Perform the request
	resp, err := http.Post(s.Base().MustJoinPathString("/bla/O1/b/O2/c/xyz").String(), "text/plain", strings.NewReader("mybody"))
	require.NoError(t, err)
	defer resp.Body.Close()

	t.Run("should return correct status code", func(t *testing.T) {
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return correctly merged headers", func(t *testing.T) {
		// Handler sets "a":"y1", filter 1's wrapper overrides with "a":"x1".
		require.Equal(t, "x1", resp.Header.Get("a"))
		// Filter 2's wrapper sets "b":"x2", but filter 1's wrapper executes last and wins, setting "b":"x1".
		require.Equal(t, "x1", resp.Header.Get("b"))
		// Filter 2's wrapper sets "c":"x2".
		require.Equal(t, "x2", resp.Header.Get("c"))
		// Handler sets "d":"y", which is untouched by filters.
		require.Equal(t, "y", resp.Header.Get("d"))
	})

	t.Run("should return body with filter modifications", func(t *testing.T) {
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		// Body should be wrapped by both filters and contain handler output.
		// Note: The original test had a non-deterministic map order. This is now fixed.
		expectedBody := "UUUU r: /bla/O1/b/O2/c/xyz - map[p1:O1 p2:O2], req: reqpre-mybody-reqsuf ZZZZ"
		require.Equal(t, expectedBody, string(data))
	})
}
