package fszjson

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/fsz/fsz"
)

type UploadOptions struct {
	Url    *urlz.Url
	Gzip   bool
	Header http.Header
}

func Upload[E any](ctx context.Context, fn func() (E, bool, error), opts UploadOptions) error {
	header := maps.Clone(opts.Header)
	if header == nil {
		header = make(http.Header)
	}
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", "application/json")
	}

	r := jsonz.LineFormatReaderStream(fn)
	defer r.Close()

	var uploadReader io.Reader = r
	if opts.Gzip {
		header.Set("Content-Encoding", "gzip")
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		_, err := io.Copy(gw, r)
		if err != nil {
			return fmt.Errorf("gzip json error: %w", err)
		}
		err = gw.Close()
		if err != nil {
			return fmt.Errorf("gzip close error: %w", err)
		}
		uploadReader = &buf
	}

	err := fsz.Upload(ctx, opts.Url, header, uploadReader)
	if err != nil {
		return fmt.Errorf("error uploading json %s: %w", opts.Url, err)
	}
	return nil
}

func UploadJson(ctx context.Context, value any, opts UploadOptions) error {
	return UploadJson(ctx, []any{value}, opts)
}

func MustUploadSlice[S ~[]E, E any](ctx context.Context, values S, opts UploadOptions) {
	err := UploadSlice(ctx, values, opts)
	errorz.Check(err)
}

func UploadSlice[S ~[]E, E any](ctx context.Context, values S, opts UploadOptions) error {
	i := 0
	return Upload(ctx, func() (E, bool, error) {
		if i >= len(values) {
			var zero E
			return zero, false, nil
		}
		val := values[i]
		i++
		return val, true, nil
	}, opts)
}
