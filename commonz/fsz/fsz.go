package fsz

import (
	"context"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/urlz"
)

var providers = map[string]FsProvider{
	"file": providerFile(),
}

func RegisterFS(scheme string, fs FsProvider) {
	providers[scheme] = fs
}

func getProvider(scheme string) (FsProvider, error) {
	prv, ok := providers[scheme]
	if !ok {
		return nil, ErrUnknownScheme
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

func Upload(ctx context.Context, url *urlz.Url, reader io.Reader) error {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return err
	}
	return p.Upload(ctx, url, reader)
}

func Stat(ctx context.Context, url *urlz.Url) (*FileStat, error) {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return nil, err
	}
	return p.Stat(ctx, url), nil
}

func Download(ctx context.Context, url *urlz.Url, callback func(found bool, headers http.Header, reader io.Reader)) error {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return err
	}
	p.Download(ctx, url, callback)
	return nil
}

func Ls(ctx context.Context, url *urlz.Url) (Paginator, error) {
	p, err := getProvider(url.Scheme)
	if err != nil {
		return nil, err
	}
	return p.Ls(ctx, url), nil
}
