package brwclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/idgen"
	"github.com/infinity6-ai/gox/commonz/ioz"
	"github.com/infinity6-ai/gox/commonz/urlz"
)

// Command represents an action sent from Go to the browser controller.
type Command struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Fn     string `json:"fn,omitempty"`
	Ref    string `json:"ref,omitempty"`
	Args   []any  `json:"args,omitempty"`
	Value  any    `json:"value,omitempty"`
}

// Reply represents the response payload sent back by the browser controller.
type Reply struct {
	ID     string          `json:"id"`
	Ok     bool            `json:"ok"`
	Result json.RawMessage `json:"result,omitempty"`
	Ref    string          `json:"ref,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// Dispatcher defines the interface for delivering commands to a browser tab.
type Dispatcher interface {
	Dispatch(ctx context.Context, tab int, cmd *Command) (*Reply, error)
}

// Ref represents a typed reference to a JavaScript object in the browser.
type Ref[T any] string

func (r Ref[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"$ref": string(r)})
}

func (r *Ref[T]) UnmarshalJSON(data []byte) error {
	var m map[string]string
	if err := json.Unmarshal(data, &m); err == nil && m["$ref"] != "" {
		*r = Ref[T](m["$ref"])
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*r = Ref[T](s)
		return nil
	}
	return fmt.Errorf("invalid ref json: %s", string(data))
}

// Client represents a handle to communicate with a browser tab.
type Client struct {
	BaseUrl    *urlz.Url
	Tab        int
	Dispatcher Dispatcher
}

// ForTab returns a shallow copy of Client targeting a different tab index.
func (c *Client) ForTab(tab int) *Client {
	return &Client{
		BaseUrl:    c.BaseUrl,
		Tab:        tab,
		Dispatcher: c.Dispatcher,
	}
}

// Create generates a unique Ref locally with zero network RPC.
func Create[T any]() Ref[T] {
	return Ref[T](idgen.B64())
}

// MustCreate generates a unique Ref locally with zero network RPC.
func MustCreate[T any]() Ref[T] {
	return Create[T]()
}

// Set binds a value in the browser tab's reference registry under the given ref.
// It returns an io.Closer that deletes the reference when closed.
func Set[T any](ctx context.Context, c *Client, ref Ref[T], v T) (io.Closer, error) {
	if c == nil || c.Dispatcher == nil {
		return nil, fmt.Errorf("dispatcher is not configured")
	}
	cmd := &Command{
		ID:     idgen.B64(),
		Action: "set",
		Ref:    string(ref),
		Value:  v,
	}
	reply, err := c.Dispatcher.Dispatch(ctx, c.Tab, cmd)
	if err != nil {
		return nil, fmt.Errorf("error setting ref %s on tab %d: %w", ref, c.Tab, err)
	}
	if !reply.Ok {
		return nil, fmt.Errorf("error setting ref %s on tab %d: %s", ref, c.Tab, reply.Error)
	}
	closer := ioz.Closer(func() error {
		return Delete(ctx, c, ref)
	})
	return closer, nil
}

// MustSet binds a value in the browser tab's reference registry, panicking on failure.
func MustSet[T any](ctx context.Context, c *Client, ref Ref[T], v T) io.Closer {
	closer, err := Set(ctx, c, ref, v)
	errorz.Check(err)
	return closer
}

// Delete deletes the reference from the browser tab's registry.
func Delete[T any](ctx context.Context, c *Client, ref Ref[T]) error {
	if c == nil || c.Dispatcher == nil {
		return fmt.Errorf("dispatcher is not configured")
	}
	cmd := &Command{
		ID:     idgen.B64(),
		Action: "delete",
		Ref:    string(ref),
	}
	reply, err := c.Dispatcher.Dispatch(ctx, c.Tab, cmd)
	if err != nil {
		return fmt.Errorf("error deleting ref %s on tab %d: %w", ref, c.Tab, err)
	}
	if !reply.Ok {
		return fmt.Errorf("error deleting ref %s on tab %d: %s", ref, c.Tab, reply.Error)
	}
	return nil
}

// MustDelete deletes the reference from the browser tab's registry, panicking on failure.
func MustDelete[T any](ctx context.Context, c *Client, ref Ref[T]) {
	err := Delete(ctx, c, ref)
	errorz.Check(err)
}

// Proc executes a procedure or JavaScript expression on the tab without expecting a return value.
func Proc(ctx context.Context, c *Client, fn string, args ...any) error {
	if c == nil || c.Dispatcher == nil {
		return fmt.Errorf("dispatcher is not configured")
	}
	cmd := &Command{
		ID:     idgen.B64(),
		Action: "proc",
		Fn:     fn,
		Args:   args,
	}
	reply, err := c.Dispatcher.Dispatch(ctx, c.Tab, cmd)
	if err != nil {
		return fmt.Errorf("error executing proc on tab %d: %w", c.Tab, err)
	}
	if !reply.Ok {
		return fmt.Errorf("javascript error in proc on tab %d: %s", c.Tab, reply.Error)
	}
	return nil
}

// MustProc executes a procedure on the tab, panicking on failure.
func MustProc(ctx context.Context, c *Client, fn string, args ...any) {
	err := Proc(ctx, c, fn, args...)
	errorz.Check(err)
}

// Func executes a function or callable expression, stores the result in the tab's reference registry,
// and returns a typed Ref[T] handle to it.
func Func[T any](ctx context.Context, c *Client, fn string, args ...any) (Ref[T], error) {
	var zero Ref[T]
	if c == nil || c.Dispatcher == nil {
		return zero, fmt.Errorf("dispatcher is not configured")
	}
	ref := Create[T]()
	cmd := &Command{
		ID:     idgen.B64(),
		Action: "func",
		Fn:     fn,
		Ref:    string(ref),
		Args:   args,
	}
	reply, err := c.Dispatcher.Dispatch(ctx, c.Tab, cmd)
	if err != nil {
		return zero, fmt.Errorf("error executing func on tab %d: %w", c.Tab, err)
	}
	if !reply.Ok {
		return zero, fmt.Errorf("javascript error in func on tab %d: %s", c.Tab, reply.Error)
	}
	return ref, nil
}

// MustFunc executes a function on the tab and returns a typed Ref handle, panicking on failure.
func MustFunc[T any](ctx context.Context, c *Client, fn string, args ...any) Ref[T] {
	ret, err := Func[T](ctx, c, fn, args...)
	errorz.Check(err)
	return ret
}

// Eval evaluates a JavaScript expression on the tab and decodes the JSON result into T.
func Eval[T any](ctx context.Context, c *Client, expr string) (T, error) {
	var zero T
	if c == nil || c.Dispatcher == nil {
		return zero, fmt.Errorf("dispatcher is not configured")
	}
	cmd := &Command{
		ID:     idgen.B64(),
		Action: "eval",
		Fn:     expr,
	}
	reply, err := c.Dispatcher.Dispatch(ctx, c.Tab, cmd)
	if err != nil {
		return zero, fmt.Errorf("error evaluating on tab %d: %w", c.Tab, err)
	}
	if !reply.Ok {
		return zero, fmt.Errorf("javascript error evaluating on tab %d: %s", c.Tab, reply.Error)
	}
	if len(reply.Result) == 0 {
		return zero, nil
	}
	var ret T
	if err := json.Unmarshal(reply.Result, &ret); err != nil {
		return zero, fmt.Errorf("failed to unmarshal eval result: %w", err)
	}
	return ret, nil
}

// MustEval evaluates a JavaScript expression on the tab, panicking on failure.
func MustEval[T any](ctx context.Context, c *Client, expr string) T {
	ret, err := Eval[T](ctx, c, expr)
	errorz.Check(err)
	return ret
}

// Get retrieves the current value of a Ref from the browser tab and decodes it into T.
func Get[T any](ctx context.Context, c *Client, ref Ref[T]) (T, error) {
	var zero T
	if c == nil || c.Dispatcher == nil {
		return zero, fmt.Errorf("dispatcher is not configured")
	}
	cmd := &Command{
		ID:     idgen.B64(),
		Action: "get",
		Ref:    string(ref),
	}
	reply, err := c.Dispatcher.Dispatch(ctx, c.Tab, cmd)
	if err != nil {
		return zero, fmt.Errorf("error getting ref %s on tab %d: %w", ref, c.Tab, err)
	}
	if !reply.Ok {
		return zero, fmt.Errorf("javascript error getting ref %s on tab %d: %s", ref, c.Tab, reply.Error)
	}
	if len(reply.Result) == 0 {
		return zero, nil
	}
	var ret T
	if err := json.Unmarshal(reply.Result, &ret); err != nil {
		return zero, fmt.Errorf("failed to unmarshal get result for ref %s: %w", ref, err)
	}
	return ret, nil
}

// MustGet retrieves the current value of a Ref, panicking on failure.
func MustGet[T any](ctx context.Context, c *Client, ref Ref[T]) T {
	ret, err := Get[T](ctx, c, ref)
	errorz.Check(err)
	return ret
}
