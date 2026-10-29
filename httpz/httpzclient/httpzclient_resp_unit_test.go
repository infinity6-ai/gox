package httpzclient_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/stretchr/testify/require"
)

func TestUnitCodeError(t *testing.T) {
	type testScenario struct {
		code    int
		payload string
		want    string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		err := httpzclient.CodeError{
			Code:    s.code,
			Payload: s.payload,
		}
		require.Equal(t, s.want, err.Error())
	}

	t.Run("formats status code and payload correctly", func(t *testing.T) {
		check(t, testScenario{
			code:    404,
			payload: "not found",
			want:    "http response error: 404, head body: 'not found'",
		})
	})

	t.Run("formats 500 with empty payload", func(t *testing.T) {
		check(t, testScenario{
			code:    500,
			payload: "",
			want:    "http response error: 500, head body: ''",
		})
	})
}

func TestUnitRespSuccess(t *testing.T) {
	type testScenario struct {
		statusCode  int
		body        string
		expectError bool
		wantCode    int
		wantPayload string
		wantErrMsg  string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		resp := &httpzclient.Resp{
			StatusCode: s.statusCode,
			Body:       io.NopCloser(strings.NewReader(s.body)),
		}
		err := resp.Success()
		if !s.expectError {
			require.NoError(t, err)
			return
		}

		require.Error(t, err)
		var codeErr httpzclient.CodeError
		require.True(t, errors.As(err, &codeErr))
		require.Equal(t, s.wantCode, codeErr.Code)
		require.Equal(t, s.wantPayload, codeErr.Payload)
		require.Equal(t, s.wantErrMsg, err.Error())
	}

	t.Run("200 OK returns nil", func(t *testing.T) {
		check(t, testScenario{
			statusCode:  http.StatusOK,
			body:        "ok",
			expectError: false,
		})
	})

	t.Run("201 Created returns nil", func(t *testing.T) {
		check(t, testScenario{
			statusCode:  http.StatusCreated,
			body:        "created",
			expectError: false,
		})
	})

	t.Run("204 No Content returns nil", func(t *testing.T) {
		check(t, testScenario{
			statusCode:  http.StatusNoContent,
			body:        "",
			expectError: false,
		})
	})

	t.Run("299 upper bound 2xx returns nil", func(t *testing.T) {
		check(t, testScenario{
			statusCode:  299,
			body:        "almost-300",
			expectError: false,
		})
	})

	t.Run("199 below 200 returns CodeError", func(t *testing.T) {
		check(t, testScenario{
			statusCode:  199,
			body:        "informational-edge",
			expectError: true,
			wantCode:    199,
			wantPayload: "informational-edge",
			wantErrMsg:  "http response error: 199, head body: 'informational-edge'",
		})
	})

	t.Run("300 Multiple Choices returns CodeError", func(t *testing.T) {
		check(t, testScenario{
			statusCode:  http.StatusMultipleChoices,
			body:        "redirect-choices",
			expectError: true,
			wantCode:    300,
			wantPayload: "redirect-choices",
			wantErrMsg:  "http response error: 300, head body: 'redirect-choices'",
		})
	})

	t.Run("404 Not Found returns CodeError", func(t *testing.T) {
		check(t, testScenario{
			statusCode:  http.StatusNotFound,
			body:        "resource-missing",
			expectError: true,
			wantCode:    404,
			wantPayload: "resource-missing",
			wantErrMsg:  "http response error: 404, head body: 'resource-missing'",
		})
	})

	t.Run("500 Internal Server Error returns CodeError", func(t *testing.T) {
		check(t, testScenario{
			statusCode:  http.StatusInternalServerError,
			body:        "server-crashed",
			expectError: true,
			wantCode:    500,
			wantPayload: "server-crashed",
			wantErrMsg:  "http response error: 500, head body: 'server-crashed'",
		})
	})

	t.Run("Body longer than 1024 bytes truncates head body to 1024 bytes", func(t *testing.T) {
		longBody := strings.Repeat("x", 2048)
		truncated := strings.Repeat("x", 1024)
		check(t, testScenario{
			statusCode:  http.StatusBadRequest,
			body:        longBody,
			expectError: true,
			wantCode:    400,
			wantPayload: truncated,
			wantErrMsg:  fmt.Sprintf("http response error: 400, head body: '%s'", truncated),
		})
	})
}

func TestUnitRespMustSuccess(t *testing.T) {
	type testScenario struct {
		statusCode  int
		body        string
		expectPanic bool
		wantPanic   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		resp := &httpzclient.Resp{
			StatusCode: s.statusCode,
			Body:       io.NopCloser(strings.NewReader(s.body)),
		}

		if s.expectPanic {
			require.PanicsWithError(t, s.wantPanic, func() {
				resp.MustSuccess()
			})
			return
		}

		require.NotPanics(t, func() {
			resp.MustSuccess()
		})
	}

	t.Run("200 OK does not panic", func(t *testing.T) {
		check(t, testScenario{
			statusCode:  http.StatusOK,
			body:        "all-good",
			expectPanic: false,
		})
	})

	t.Run("404 Not Found panics with structured error", func(t *testing.T) {
		check(t, testScenario{
			statusCode:  http.StatusNotFound,
			body:        "page-missing",
			expectPanic: true,
			wantPanic:   "http response error: 404, head body: 'page-missing': (InternalError, code=500)",
		})
	})
}
