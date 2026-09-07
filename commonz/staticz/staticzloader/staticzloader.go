package staticzloader

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"sync"

	"github.com/infinity6-ai/gox/commonz/encz/enczb64"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/filez"
)

type Codes struct {
	codes map[any][]byte
	mu    sync.RWMutex
}

var codes = &Codes{
	codes: make(map[any][]byte),
}

func SetCode(name any, tgz func() string) {
	decoded := decode(tgz)
	codes.mu.Lock()
	defer codes.mu.Unlock()
	codes.codes[name] = decoded
}

func decode(tgz func() string) []byte {
	encoded := tgz()
	b, err := enczb64.StdDecode(encoded)
	errorz.Checkf(err, "error decoding")
	return b.Bytes()
}

func GetCode(name any) []byte {
	codes.mu.RLock()
	defer codes.mu.RUnlock()
	return codes.codes[name]
}

func Walk(ctx context.Context, name any, callback func(entry filez.WalkLoaderEntry) error) error {
	code := GetCode(name)
	if code == nil {
		return fmt.Errorf("code not found: %s (%T)", name, name)
	}

	gzReader, err := gzip.NewReader(bytes.NewBuffer(code))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var skipPrefix string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading from tar archive: %w", err)
		}

		if skipPrefix != "" && strings.HasPrefix(header.Name, skipPrefix) {
			continue
		}
		skipPrefix = ""

		fileInfo := header.FileInfo()
		d := filez.NewWalkLoaderEntry(header.Name, fileInfo.Name(), fileInfo.IsDir(), fileInfo.Size(), func() (io.ReadCloser, error) {
			if fileInfo.IsDir() {
				return nil, fmt.Errorf("cannot WalkLoad a directory: %s", header.Name)
			}
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("error reading file data: %w", err)
			}
			return io.NopCloser(bytes.NewBuffer(data)), nil
		})

		err = callback(d)
		if err != nil {
			if err == fs.SkipDir && d.IsDir() {
				skipPrefix = header.Name
				if !strings.HasSuffix(skipPrefix, "/") {
					skipPrefix += "/"
				}
				continue
			}
			if err == fs.SkipAll {
				return nil
			}
			return err
		}
	}
	return nil
}
