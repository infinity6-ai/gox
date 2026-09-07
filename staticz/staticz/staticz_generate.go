package staticz

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"encoding/base64"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/infinity6-ai/gox/staticz/internal"
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
		panic("imp is required")
	}
	if g.Code == "" {
		panic("code is required")
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
	log.Printf("src: %s", srcDir)

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
	opts.fix()

	w := bufio.NewWriter(opts.Out)

	/*
			7 09:23:08 i6dev [cmd_go-platform_static] echo 'package stfiles'
		2026-09-07 09:23:08 i6dev [cmd_go-platform_static] echo 'import "go.code.infinity6.ai/platform/httpz/statichandler"'
		2026-09-07 09:23:08 i6dev [cmd_go-platform_static] echo 'func init() {'
		2026-09-07 09:23:08 i6dev [cmd_go-platform_static] echo '    statichandler.Instance().SetCode(pack, func() string {'
		2026-09-07 09:23:08 i6dev [cmd_go-platform_static] echo '        return `'
		2026-09-07 09:23:08 i6dev [cmd_go-platform_static] base64
		2026-09-07 09:23:08 i6dev [cmd_go-platform_static] tar czf - -C . static
		2026-09-07 09:23:08 i6dev [cmd_go-platform_static] echo '        `'
		2026-09-07 09:23:08 i6dev [cmd_go-platform_static] echo '   })'
		2026-09-07 09:23:08 i6dev [cmd_go-platform_static] echo '}'
	*/

	w.WriteString("package stzfiles\n")
	w.WriteString("import \"")
	w.WriteString(opts.Imp)
	w.WriteString("\"\n")
	w.WriteString("func init() {\n")
	w.WriteString("    ")
	w.WriteString(opts.Code)
	w.WriteString(" {\n")
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
