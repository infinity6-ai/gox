package brwclient_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
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

func TestUnitRefJson(t *testing.T) {
	type testScenario struct {
		name     string
		ref      brwclient.Ref[string]
		expected string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		data, err := json.Marshal(s.ref)
		require.NoError(t, err)
		require.Equal(t, s.expected, string(data))

		var unmarshaled brwclient.Ref[string]
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		require.Equal(t, s.ref, unmarshaled)
	}

	t.Run("Ref marshals to ref envelope", func(t *testing.T) {
		check(t, testScenario{
			name:     "standard ref",
			ref:      brwclient.Ref[string]("ref-12345"),
			expected: `{"$ref":"ref-12345"}`,
		})
	})
}

func TestUnitClientOperations(t *testing.T) {
	type testScenario struct {
		name       string
		action     func(ctx context.Context, c *brwclient.Client) error
		wantAction string
		wantRef    string
		wantErr    string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		var capturedCmd *brwclient.Command
		mock := &mockDispatcher{
			dispatchFunc: func(ctx context.Context, cmd *brwclient.Command) (*brwclient.Reply, error) {
				capturedCmd = cmd
				if s.wantErr != "" && s.wantErr == "dispatch_failed" {
					return nil, fmt.Errorf("network error")
				}
				if s.wantErr != "" && s.wantErr == "js_error" {
					return &brwclient.Reply{ID: cmd.ID, Ok: false, Error: "ReferenceError: foo is not defined"}, nil
				}
				var rawRes json.RawMessage
				if cmd.Action == "eval" || cmd.Action == "get" || cmd.Action == "call" {
					rawRes = json.RawMessage(`"hello world"`)
				}
				return &brwclient.Reply{
					ID:     cmd.ID,
					Ok:     true,
					Result: rawRes,
					Ref:    cmd.Ref,
				}, nil
			},
		}

		c := &brwclient.Client{Dispatcher: mock}
		err := s.action(t.Context(), c)

		if s.wantErr != "" {
			require.Error(t, err)
			return
		}

		require.NoError(t, err)
		if s.wantAction != "" {
			require.NotNil(t, capturedCmd)
			require.Equal(t, s.wantAction, capturedCmd.Action)
		}
	}

	t.Run("Proc executes command without error", func(t *testing.T) {
		check(t, testScenario{
			name: "proc success",
			action: func(ctx context.Context, c *brwclient.Client) error {
				return brwclient.Proc(ctx, c, "console.log", "hello")
			},
			wantAction: "proc",
		})
	})

	t.Run("Proc propagates JS error", func(t *testing.T) {
		check(t, testScenario{
			name: "proc js error",
			action: func(ctx context.Context, c *brwclient.Client) error {
				return brwclient.Proc(ctx, c, "invalid.syntax()")
			},
			wantErr: "js_error",
		})
	})

	t.Run("Eval returns decoded value", func(t *testing.T) {
		check(t, testScenario{
			name: "eval success",
			action: func(ctx context.Context, c *brwclient.Client) error {
				val, err := brwclient.Eval[string](ctx, c, `"hello " + "world"`)
				if err != nil {
					return err
				}
				require.Equal(t, "hello world", val)
				return nil
			},
			wantAction: "eval",
		})
	})

	t.Run("Call executes function with arguments and returns value", func(t *testing.T) {
		check(t, testScenario{
			name: "call success",
			action: func(ctx context.Context, c *brwclient.Client) error {
				val, err := brwclient.Call[string](ctx, c, `(a, b) => a + " " + b`, "hello", "world")
				if err != nil {
					return err
				}
				require.Equal(t, "hello world", val)
				return nil
			},
			wantAction: "call",
		})
	})

	t.Run("Func creates reference", func(t *testing.T) {
		check(t, testScenario{
			name: "func success",
			action: func(ctx context.Context, c *brwclient.Client) error {
				ref, err := brwclient.Func[string](ctx, c, "() => 'res'")
				if err != nil {
					return err
				}
				require.NotEmpty(t, ref)
				return nil
			},
			wantAction: "func",
		})
	})

	t.Run("Set and Closer deletes reference", func(t *testing.T) {
		deletedRef := ""
		mock := &mockDispatcher{
			dispatchFunc: func(ctx context.Context, cmd *brwclient.Command) (*brwclient.Reply, error) {
				if cmd.Action == "delete" {
					deletedRef = cmd.Ref
				}
				return &brwclient.Reply{ID: cmd.ID, Ok: true}, nil
			},
		}

		c := &brwclient.Client{Dispatcher: mock}
		ref := brwclient.MustCreate[string]()
		closer, err := brwclient.Set(t.Context(), c, ref, "test-value")
		require.NoError(t, err)
		require.NotNil(t, closer)

		err = closer.Close()
		require.NoError(t, err)
		require.Equal(t, string(ref), deletedRef)
	})

	t.Run("RefDump returns active references map", func(t *testing.T) {
		mock := &mockDispatcher{
			dispatchFunc: func(ctx context.Context, cmd *brwclient.Command) (*brwclient.Reply, error) {
				return &brwclient.Reply{
					ID:     cmd.ID,
					Ok:     true,
					Result: json.RawMessage(`{"ref-1":"HTMLCanvasElement","ref-2":"Function"}`),
				}, nil
			},
		}

		c := &brwclient.Client{Dispatcher: mock}
		dump, err := brwclient.RefDump(t.Context(), c)
		require.NoError(t, err)
		require.Equal(t, "HTMLCanvasElement", dump["ref-1"])
		require.Equal(t, "Function", dump["ref-2"])

		dumpSelective, err := brwclient.RefDump(t.Context(), c, "ref-1")
		require.NoError(t, err)
		require.Equal(t, "HTMLCanvasElement", dumpSelective["ref-1"])
	})
}
