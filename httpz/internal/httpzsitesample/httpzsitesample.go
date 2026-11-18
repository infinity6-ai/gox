// Package httpzsitesample provides a sample static site implementation for demonstration and CLI commands.
package httpzsitesample

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"

	"github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
	"github.com/infinity6-ai/gox/httpz/httpzsite"
)

type tpack string

// PackName is the staticz pack name registered for the sample site.
const PackName tpack = "sample-site-pack"

func init() {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	files := map[string]string{
		"index.html":     "<!DOCTYPE html><html><head><title>Sample Site</title><link rel=\"stylesheet\" href=\"style.css\"></head><body><h1>Hello from Sample Site!</h1></body></html>",
		"style.css":      "body { font-family: sans-serif; margin: 2rem; color: #333; }\nh1 { color: #0066cc; }",
		"sub/index.html": "<!DOCTYPE html><html><body><h2>Sub page</h2></body></html>",
	}

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		_ = tw.WriteHeader(hdr)
		_, _ = tw.Write([]byte(content))
	}
	_ = tw.Close()
	_ = gw.Close()

	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	staticzloader.SetCode(PackName, func() string { return b64 })
}

// New creates and returns a new sample Site instance.
func New() *httpzsite.Site {
	return &httpzsite.Site{
		LocalAddress: "localhost:8080",
		Pack:         PackName,
	}
}
