package staticzloader

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
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

func Walk(ctx context.Context, name any) {
	code := GetCode(name)
	if code == nil {
		panic(fmt.Sprintf("code not find: %s (%T)", name, name))
	}

	gzReader, err := gzip.NewReader(bytes.NewBuffer(code))
	errorz.Check(err)
	defer gzReader.Close()

	// 3. Create a tar reader wrapping the gzip reader
	tarReader := tar.NewReader(gzReader)

	// 4. Iterate over the files in the archive
	for {
		header, err := tarReader.Next()

		// io.EOF means we reached the end of the archive
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading tar: %v\n", err)
			return
		}

		// Check the type of the entry
		switch header.Typeflag {
		case tar.TypeDir:
			// directory, ignore it
		case tar.TypeReg:
			logger.Info(ctx, "File", map[string]any{"name": header.Name, "size": header.Size})
			// Optional: Read the file content without writing to disk
			// content, _ := io.ReadAll(tarReader)
		default:
			panic(fmt.Sprintf("unsupported type %s (%T) %s: %v", name, name, header.Name, header.Typeflag))
		}
	}
}
