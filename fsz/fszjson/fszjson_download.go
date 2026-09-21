package fszjson

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/fsz/fsz"
)

type DownloadOptions struct {
	Url  *urlz.Url
	Gzip bool
}

func Download[E any](ctx context.Context, fn func(item E) (bool, error), opts DownloadOptions) (bool, http.Header, error) {
	var found bool
	var header http.Header
	err := fsz.Download(ctx, opts.Url, func(f bool, h http.Header, r io.Reader) error {
		found = f
		if !found {
			return nil
		}
		header = h

		var reader io.Reader = r
		if opts.Gzip {
			gr, err := gzip.NewReader(r)
			if err != nil {
				return fmt.Errorf("gunzip json error: %w", err)
			}
			defer gr.Close()
			reader = gr
		}

		err := jsonz.LineParseStream(reader, fn)
		if err != nil {
			return fmt.Errorf("error parsing json stream from %s: %w", opts.Url, err)
		}
		return nil
	})
	if err != nil {
		return false, nil, fmt.Errorf("error downloading json from %s: %w", opts.Url, err)
	}
	return found, header, nil
}

func DownloadSlice[S ~[]E, E any](ctx context.Context, v *S, opts DownloadOptions) (bool, http.Header, error) {
	var results S
	found, header, err := Download(ctx, func(item E) (bool, error) {
		results = append(results, item)
		return true, nil
	}, opts)
	if err == nil && found {
		*v = results
	}
	return found, header, err
}

func MustDownloadJson[T any](ctx context.Context, url *urlz.Url, v T) (f bool, h http.Header) {
	var err error
	f, h, err = DownloadJson(ctx, url, v)
	errorz.Check(err)
	return
}

func DownloadJson[T any](ctx context.Context, url *urlz.Url, v T) (f bool, h http.Header, e error) {
	e = fsz.Download(ctx, url, func(found bool, headers http.Header, reader io.Reader) error {
		f = found
		if !f {
			return nil
		}
		h = headers
		_, err := jsonz.ParseReader(reader, v)
		if err != nil {
			return fmt.Errorf("error downloading json: %w", err)
		}
		return nil
	})
	return
}
