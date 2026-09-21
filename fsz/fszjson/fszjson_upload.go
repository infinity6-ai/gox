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
	Values []any
	Gzip   bool
	Header http.Header
}

func Upload(ctx context.Context, opts UploadOptions) error {
	header := maps.Clone(opts.Header)
	if header == nil {
		header = make(http.Header)
	}
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", "application/json")
	}
	if len(opts.Values) > 1 {
		panic("not implement yet")
	}
	r := jsonz.FormatReadCloser(opts.Values[0])
	defer r.Close()
	var err error
	if opts.Gzip {
		r, err = gzip.NewReader(r)
		if err != nil {
			return fmt.Errorf("gunzip json error: %w", err)
		}
	}
	err = fsz.Upload(ctx, opts.Url, header, r)
	return fmt.Errorf("error uploading json %s: %w", opts.Url, err)
}
