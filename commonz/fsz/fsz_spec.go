package fsz

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/infinity6-ai/gox/commonz/urlz"
)

var ErrUnknownScheme = errors.New("unknown scheme")
var ErrUnsupportedOperation = errors.New("unsupported operation")

type FileStat struct {
	Url         urlz.Url   `json:"url"`
	ContentType string     `json:"content_type"`
	Md5         string     `json:"md5"`
	Size        uint64     `json:"size"`
	Etag        string     `json:"etag"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type Paginator interface {
	Paginate(ctx context.Context, max int) []*FileStat
}

type FsProvider interface {
	Stat(ctx context.Context, url *urlz.Url) *FileStat
	Upload(ctx context.Context, url *urlz.Url, reader io.Reader) error
	Download(ctx context.Context, url *urlz.Url, callback func(found bool, headers http.Header, reader io.Reader))
	Delete(ctx context.Context, url *urlz.Url) error

	Ls(ctx context.Context, prefix *urlz.Url) Paginator

	SignGet(ctx context.Context, url *urlz.Url, duration time.Duration) (string, error)
	SignPut(ctx context.Context, url *urlz.Url, duration time.Duration) (string, error)
	SignDelete(ctx context.Context, url *urlz.Url, duration time.Duration) (string, error)
}
