package fszjson

import (
	"compress/gzip"
	"context"
	"fmt"
	"maps"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/fsz/fsz"
)

type UploadOptions struct {
	Url    *urlz.Url
	Gzip   bool
	Header http.Header
}

func Upload[S ~[]E, E any](ctx context.Context, values S, opts UploadOptions) error {
	header := maps.Clone(opts.Header)
	if header == nil {
		header = make(http.Header)
	}
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", "application/json")
	}
	r := jsonz.LineFormatReadCloser(values)
	defer r.Close()
	var err error
	if opts.Gzip {
		r, err = gzip.NewReader(r)
		if err != nil {
			return fmt.Errorf("gunzip json error: %w", err)
		}
		defer r.Close()
	}
	err = fsz.Upload(ctx, opts.Url, header, r)
	if err != nil {
		return fmt.Errorf("error uploading json %s: %w", opts.Url, err)
	}
	return nil
}
