package fszjson

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/fsz/fsz"
)

type DownloadOptions struct {
	Url  *urlz.Url
	Gzip bool
}

func Download[S ~[]E, E any](ctx context.Context, v *S, opts DownloadOptions) (bool, http.Header, error) {
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

		err := jsonz.LineParseReader(reader, v)
		if err != nil {
			return fmt.Errorf("error parsing json lines from %s: %w", opts.Url, err)
		}
		return nil
	})
	if err != nil {
		return false, nil, fmt.Errorf("error downloading json from %s: %w", opts.Url, err)
	}
	return found, header, nil
}
