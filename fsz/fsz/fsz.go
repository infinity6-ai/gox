package fsz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/infinity6-ai/gox/commonz/deferz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/fsz/fsz/fszsignmethod"
)

var providers = map[string]FsProvider{
	"file": providerFile(),
	"gs":   providerGs(),
}

func RegisterFS(ctx context.Context, scheme string, fs FsProvider) io.Closer {
	dfz := deferz.New(ctx)
	defer dfz.Close()

	old := providers[scheme]
	if old != nil {
		dfz.Add(func() {
			providers[scheme] = old
		})
	}
	providers[scheme] = fs
	return dfz.Detach()
}

func getProvider(scheme string) (FsProvider, error) {
	prv, ok := providers[scheme]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownScheme, scheme)
	}
	return prv, nil
}

func MustDelete(ctx context.Context, url *urlz.Url) {
	err := Delete(ctx, url)
	errorz.Check(err)
}

func Delete(ctx context.Context, url *urlz.Url) error {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return err
	}
	return p.Delete(ctx, url)
}

func MustUpload(ctx context.Context, url *urlz.Url, headers http.Header, reader io.Reader) {
	errorz.Check(Upload(ctx, url, headers, reader))
}

func Upload(ctx context.Context, url *urlz.Url, headers http.Header, reader io.Reader) error {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return err
	}
	return p.Upload(ctx, url, headers, reader)
}

func Stat(ctx context.Context, url *urlz.Url) (*FileStat, error) {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return nil, err
	}
	return p.Stat(ctx, url)
}

func MustDownload(ctx context.Context, url *urlz.Url, callback func(found bool, headers http.Header, reader io.Reader)) {
	err := Download(ctx, url, func(found bool, headers http.Header, reader io.Reader) error {
		callback(found, headers, reader)
		return nil
	})
	errorz.Check(err)
}

func Download(ctx context.Context, url *urlz.Url, callback func(found bool, headers http.Header, reader io.Reader) error) error {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return err
	}
	return p.Download(ctx, url, callback)
}

type walkerPaginator struct {
	fn     func(ctx context.Context, url *urlz.Url, walker Walker) error
	url    *urlz.Url
	cursor string
	done   bool
}

func (p *walkerPaginator) GetCursor() string {
	return p.cursor
}

func (p *walkerPaginator) SetCursor(cursor string) {
	p.cursor = cursor
	p.done = false
}

func (p *walkerPaginator) Paginate(ctx context.Context, max int) ([]*FileStat, error) {
	if max <= 0 || p.done {
		return nil, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error during paginate: %w", err)
	}

	var results []*FileStat
	called := false
	err := p.fn(ctx, p.url, Walker{
		StartCursor: p.cursor,
		PageSize:    max,
		Pager: func(page []*FileStat, nextCursor string) error {
			called = true
			results = page
			p.cursor = nextCursor
			if nextCursor == "" {
				p.done = true
			}
			return ErrStop
		},
	})
	if err != nil && !errors.Is(err, ErrStop) {
		return nil, err
	}
	if !called || p.cursor == "" {
		p.done = true
	}
	return results, nil
}

func Ls(ctx context.Context, url *urlz.Url) (Paginator, error) {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return nil, err
	}
	return &walkerPaginator{
		fn:  p.Lister,
		url: url,
	}, nil
}

func Find(ctx context.Context, url *urlz.Url) (Paginator, error) {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return nil, err
	}
	return &walkerPaginator{
		fn:  p.Finder,
		url: url,
	}, nil
}

func Lister(ctx context.Context, url *urlz.Url, walker Walker) error {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return err
	}
	return p.Lister(ctx, url, walker)
}

func Finder(ctx context.Context, url *urlz.Url, walker Walker) error {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return err
	}
	return p.Finder(ctx, url, walker)
}

func Copy(ctx context.Context, src *urlz.Url, dest *urlz.Url) error {
	if src.Scheme != dest.Scheme {
		srcProvider, err := getProvider(src.Scheme)
		if err != nil {
			return err
		}
		destProvider, err := getProvider(dest.Scheme)
		if err != nil {
			return err
		}

		return srcProvider.Download(ctx, src, func(found bool, headers http.Header, reader io.Reader) error {
			// headers = CopyHeaders(headers, nil)
			return destProvider.Upload(ctx, dest, headers, reader)
		})
	}

	prv, err := getProvider(src.Scheme)
	if err != nil {
		return err
	}
	return prv.Copy(ctx, src, dest)
}

func MustSign(ctx context.Context, signMethod fszsignmethod.SignMethod, url *urlz.Url, duration time.Duration) string {
	ret, err := Sign(ctx, signMethod, url, duration)
	errorz.Check(err)
	return ret
}

func Sign(ctx context.Context, signMethod fszsignmethod.SignMethod, url *urlz.Url, duration time.Duration) (string, error) {
	prv, err := getProvider(url.Scheme)
	if err != nil {
		return "", err
	}
	return prv.Sign(ctx, signMethod, url, duration)
}

func RmTree(ctx context.Context, url *urlz.Url) error {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return err
	}
	return p.RmTree(ctx, url)
}

func Move(ctx context.Context, src *urlz.Url, dest *urlz.Url) error {
	if src.Scheme != dest.Scheme {
		return fmt.Errorf("move between different schemes is not supported (%s -> %s)", src.Scheme, dest.Scheme)
	}
	p, err := getProvider(src.Scheme)
	if err != nil {
		return err
	}
	return p.Move(ctx, src, dest)
}
