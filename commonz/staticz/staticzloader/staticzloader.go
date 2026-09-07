package staticzloader

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sync"

	"github.com/infinity6-ai/gox/commonz/encz/enczb64"
	"github.com/infinity6-ai/gox/commonz/errorz"
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
	errorz.Check(fmt.Errorf("%w: error decoding", err))
	return b.Bytes()
}

func GetCode(name any) []byte {
	codes.mu.RLock()
	defer codes.mu.RUnlock()
	return codes.codes[name]
}

func Walk(name any) {
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
			fmt.Printf("Directory: %s\n", header.Name)
		case tar.TypeReg:
			fmt.Printf("File: %s (%d bytes)\n", header.Name, header.Size)
			// Optional: Read the file content without writing to disk
			// content, _ := io.ReadAll(tarReader)
		default:
			fmt.Printf("Other entry: %s (Type: %c)\n", header.Name, header.Typeflag)
		}
	}
}
