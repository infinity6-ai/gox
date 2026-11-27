package httpzsite_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"

	"github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
)

type tpack string

const testPackName tpack = "test-pack-site"

func createSampleTgz(files map[string]string) func() string {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			panic(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			panic(err)
		}
	}
	_ = tw.Close()
	_ = gw.Close()
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return func() string { return b64 }
}

func init() {
	sampleTgz := createSampleTgz(map[string]string{
		"index.html":        "<html><body>Home</body></html>",
		"style.css":         "body { color: red; }",
		"sub/index.html":    "<html><body>Sub</body></html>",
		"sub/script.js":     "console.log('sub');",
		"deep/nested/a.txt": "deep-content",
	})
	staticzloader.SetCode(testPackName, sampleTgz)
}
