package statizwriter

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

type GenerateOptions struct {
	Imp  string
	Code string
	Name string
	Dir  string
	Out  io.Writer
}

func (g *GenerateOptions) fix() {
	if g.Imp == "" {
		g.Imp = "github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
	}
	if g.Code == "" {
		g.Code = "staticzloader.SetCode"
	}
	if g.Dir == "" {
		g.Dir = "stzfiles"
	}
	if g.Out == nil {
		g.Out = os.Stdout
	}
}

func CreateTarGz(srcDir string, out io.Writer) {
	gw := gzip.NewWriter(out)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	srcDir = filepath.Clean(srcDir)

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		errorz.Check(err)

		header, err := tar.FileInfoHeader(info, info.Name())
		errorz.Check(err)

		relPath, err := filepath.Rel(srcDir, path)
		errorz.Check(err)

		if relPath == "." {
			return nil
		}

		header.Name = filepath.ToSlash(relPath)

		err = tw.WriteHeader(header)
		errorz.Check(err)

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		errorz.Check(err)
		defer file.Close()

		_, err = io.Copy(tw, file)
		errorz.Check(err)

		return nil
	})

	errorz.Check(err)
}

func Generate(ctx context.Context, opts GenerateOptions) {
	opts.fix()

	w := bufio.NewWriter(opts.Out)

	w.WriteString("package stzfiles\n")
	w.WriteString("import \"")
	w.WriteString(opts.Imp)
	w.WriteString("\"\n")
	w.WriteString("func init() {\n")
	w.WriteString("    ")
	w.WriteString(opts.Code)
	w.WriteString("(Name, func() string {\n")
	w.WriteString("        return `\n")

	enc := base64.StdEncoding
	wcols := &lineEncoder{w: w, limit: 80}
	b64 := base64.NewEncoder(enc, wcols)
	CreateTarGz(opts.Dir, b64)
	b64.Close()

	w.WriteString("\n        `\n")
	w.WriteString("   })\n")
	w.WriteString("}\n")

	w.Flush()
}
