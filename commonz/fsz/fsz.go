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
