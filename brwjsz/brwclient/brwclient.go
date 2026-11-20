package brwclient

import (
	"context"

	"github.com/infinity6-ai/gox/commonz/urlz"
)

type Client struct {
	BaseUrl *urlz.Url
}

type Ref[T any] string

func MustSet[T any](ctx context.Context, c *Client, ref Ref[T], v T) {
	panic("implement")
}

func MustDelete[T any](ctx context.Context, c *Client, ref Ref[T]) {
	panic("implement")
}

func MustCreate[T any](ctx context.Context, c *Client, v T) Ref[T] {
	panic("implement")
}

func MustProc(ctx context.Context, c *Client, fn string, args ...any) {
	panic("implement")
}

func MustFunc[T any](ctx context.Context, c *Client, fn string, args ...any) Ref[T] {
	panic("implement")
}
