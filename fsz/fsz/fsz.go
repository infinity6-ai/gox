package fsz

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/infinity6-ai/gox/commonz/deferz"
	"github.com/infinity6-ai/gox/commonz/urlz"
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

func Delete(ctx context.Context, url *urlz.Url) error {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return err
	}
	return p.Delete(ctx, url)
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

func Download(ctx context.Context, url *urlz.Url, callback func(found bool, headers http.Header, reader io.Reader) error) error {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return err
	}
	return p.Download(ctx, url, callback)
}

func Ls(ctx context.Context, url *urlz.Url) (Paginator, error) {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return nil, err
	}
	return p.Ls(ctx, url)
}

func Find(ctx context.Context, url *urlz.Url) (Paginator, error) {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return nil, err
	}
	return p.Find(ctx, url)
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

func SignGet(ctx context.Context, url *urlz.Url, duration time.Duration) (string, error) {
	prv, err := getProvider(url.Scheme)
	if err != nil {
		return "", err
	}
	return prv.SignGet(ctx, url, duration)
}

func SignPut(ctx context.Context, url *urlz.Url, duration time.Duration) (string, error) {
	prv, err := getProvider(url.Scheme)
	if err != nil {
		return "", err
	}
	return prv.SignPut(ctx, url, duration)
}

func SignDelete(ctx context.Context, url *urlz.Url, duration time.Duration) (string, error) {
	prv, err := getProvider(url.Scheme)
	if err != nil {
		return "", err
	}
	return prv.SignDelete(ctx, url, duration)
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

