package brwqunit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/brwjsz"
	"github.com/infinity6-ai/gox/brwjsz/brwqunit"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

type mockDispatcher struct {
	dispatchFunc func(ctx context.Context, cmd *brwclient.Command) (*brwclient.Reply, error)
}

func (m *mockDispatcher) Dispatch(ctx context.Context, cmd *brwclient.Command) (*brwclient.Reply, error) {
	if m.dispatchFunc != nil {
		return m.dispatchFunc(ctx, cmd)
	}
	return &brwclient.Reply{ID: cmd.ID, Ok: true}, nil
}

func TestUnitRunMock(t *testing.T) {
	mock := &mockDispatcher{
		dispatchFunc: func(ctx context.Context, cmd *brwclient.Command) (*brwclient.Reply, error) {
			if cmd.Action == "call" {
				report := &brwqunit.TestReport{
					Status: "passed",
					Total:  2,
					Passed: 2,
					Tests: []brwqunit.TestResult{
						{Name: "test1", SuiteName: "suite", Status: "passed"},
						{Name: "test2", SuiteName: "suite", Status: "passed"},
					},
				}
				data, _ := json.Marshal(report)
				return &brwclient.Reply{ID: cmd.ID, Ok: true, Result: data}, nil
			}
			return &brwclient.Reply{ID: cmd.ID, Ok: true}, nil
		},
	}

	c := &brwclient.Client{Dispatcher: mock}
	report, err := brwqunit.Run(t.Context(), c, "/tests/test.js")
	require.NoError(t, err)
	require.Equal(t, "passed", report.Status)
	require.Equal(t, 2, report.Total)
	require.Equal(t, 2, report.Passed)
}

func TestUnitRunE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()

	opts := brwjsz.Options{}
	opts.Chrome("", true, "")

	b := brwjsz.New(ctx, opts)
	defer b.Close()

	// Register sample passing test script
	mathTestJs := `
		QUnit.module('MathSuite', () => {
			QUnit.test('addition', (assert) => {
				assert.equal(2 + 3, 5, '2 + 3 should be 5');
			});
			QUnit.test('multiplication', (assert) => {
				assert.equal(4 * 5, 20, '4 * 5 should be 20');
			});
		});
	`
	b.Server().AddHandler("GET", "/tests/math-test.js", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		headers := make(http.Header)
		headers.Set("Content-Type", "application/javascript")
		w := resp(http.StatusOK, headers)
		_, _ = w.Write([]byte(mathTestJs))
	})

	// Register sample failing test script
	failingTestJs := `
		QUnit.module('FailSuite', () => {
			QUnit.test('expected failure', (assert) => {
				assert.equal(1 + 1, 3, 'deliberate mismatch');
			});
		});
	`
	b.Server().AddHandler("GET", "/tests/fail-test.js", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		headers := make(http.Header)
		headers.Set("Content-Type", "application/javascript")
		w := resp(http.StatusOK, headers)
		_, _ = w.Write([]byte(failingTestJs))
	})

	err := b.Start(ctx)
	require.NoError(t, err)

	client := b.Client()

	// 1. Run passing tests
	report, err := brwqunit.Run(ctx, client, "/tests/math-test.js")
	require.NoError(t, err)
	require.Equal(t, "passed", report.Status)
	require.Equal(t, 2, report.Total)
	require.Equal(t, 2, report.Passed)
	require.Equal(t, 0, report.Failed)

	// 2. Run passing tests via brwqunit.Test (maps to t.Run)
	brwqunit.Test(t, client, "/tests/math-test.js")

	// 3. Run failing test and inspect structured error
	failReport, err := brwqunit.Run(ctx, client, "/tests/fail-test.js")
	require.NoError(t, err)
	require.Equal(t, "failed", failReport.Status)
	require.Equal(t, 1, failReport.Total)
	require.Equal(t, 1, failReport.Failed)
	require.Len(t, failReport.Tests, 1)
	require.Equal(t, "expected failure", failReport.Tests[0].Name)
	require.NotEmpty(t, failReport.Tests[0].Errors)
	require.Contains(t, failReport.Tests[0].Errors[0].Message, "deliberate mismatch")
}
