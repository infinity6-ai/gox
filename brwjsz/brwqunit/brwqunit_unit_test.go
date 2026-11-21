package brwqunit_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/brwqunit"
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

func TestUnitRunWithOptionsMock(t *testing.T) {
	mock := &mockDispatcher{
		dispatchFunc: func(ctx context.Context, cmd *brwclient.Command) (*brwclient.Reply, error) {
			if cmd.Action == "call" {
				report := &brwqunit.TestReport{
					Status: "passed",
					Total:  1,
					Passed: 1,
					Tests: []brwqunit.TestResult{
						{Name: "importmap-test", SuiteName: "suite", Status: "passed"},
					},
				}
				data, _ := json.Marshal(report)
				return &brwclient.Reply{ID: cmd.ID, Ok: true, Result: data}, nil
			}
			return &brwclient.Reply{ID: cmd.ID, Ok: true}, nil
		},
	}

	c := &brwclient.Client{Dispatcher: mock}
	report, err := brwqunit.RunWithOptions(t.Context(), c, brwqunit.Options{
		TestUrls:  []string{"/tests/test.js"},
		ImportMap: map[string]string{"foo": "/bar.js"},
	})
	require.NoError(t, err)
	require.Equal(t, "passed", report.Status)
	require.Equal(t, 1, report.Total)
	require.Equal(t, 1, report.Passed)
}
