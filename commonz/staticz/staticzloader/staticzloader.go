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
	"github.com/infinity6-ai/gox/commonz/logz"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

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

// dirEntry implements fs.DirEntry for a fs.FileInfo.
type dirEntry struct {
	fs.FileInfo
}

// Type returns the file mode type of the file.
func (d dirEntry) Type() fs.FileMode {
	return d.Mode().Type()
}

// Info returns the fs.FileInfo for the file.
func (d dirEntry) Info() (fs.FileInfo, error) {
	return d.FileInfo, nil
}

func Walk(ctx context.Context, name any, fn fs.WalkDirFunc) error {
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
		d := dirEntry{fileInfo}
		err = fn(header.Name, d, nil)
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
