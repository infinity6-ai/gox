package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
)

type GenerateOptions struct {
	Imp  string
	Code string
}

func CreateTarGz(srcDir string, out io.Writer) {
	gw := gzip.NewWriter(out)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	srcDir = filepath.Clean(srcDir)

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		check(err)

		header, err := tar.FileInfoHeader(info, info.Name())
		check(err)

		relPath, err := filepath.Rel(srcDir, path)
		check(err)

		if relPath == "." {
			return nil
		}

		header.Name = filepath.ToSlash(relPath)

		err = tw.WriteHeader(header)
		check(err)

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		check(err)
		defer file.Close()

		_, err = io.Copy(tw, file)
		check(err)

		return nil
	})

	check(err)
}

func Generate(ctx context.Context, opts GenerateOptions) {

}
