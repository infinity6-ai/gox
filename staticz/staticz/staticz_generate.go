package staticz

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/infinity6-ai/gox/staticz/internal"
)

type GenerateOptions struct {
	Imp  string
	Code string
	Out  io.Writer
}

func CreateTarGz(srcDir string, out io.Writer) {
	gw := gzip.NewWriter(out)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	srcDir = filepath.Clean(srcDir)

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		internal.Check(err)

		header, err := tar.FileInfoHeader(info, info.Name())
		internal.Check(err)

		relPath, err := filepath.Rel(srcDir, path)
		internal.Check(err)

		if relPath == "." {
			return nil
		}

		header.Name = filepath.ToSlash(relPath)

		err = tw.WriteHeader(header)
		internal.Check(err)

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		internal.Check(err)
		defer file.Close()

		_, err = io.Copy(tw, file)
		internal.Check(err)

		return nil
	})

	internal.Check(err)
}

func Generate(ctx context.Context, opts GenerateOptions) {

}
