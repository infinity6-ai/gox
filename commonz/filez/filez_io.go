package filez

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/strconvz"
)

// -----------------------------------------------------------------------------
// Writing & Moving
// -----------------------------------------------------------------------------

// Write writes the given payload to a destination. The destination can be a file
// path or a file descriptor specified with the `@` prefix (e.g., `@1` for stdout).
// It panics if any error occurs.
func Write(dest string, payload []byte) {
	fd := -1
	if strings.HasPrefix(dest, "@") {
		fdStr := strings.TrimPrefix(dest, "@")
		fd = strconvz.MustParseNumber[int](fdStr)
	} else {
		file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		errorz.Check(err)
		defer file.Close()
		fd = int(file.Fd())
	}
	_, err := syscall.Write(fd, payload)
	errorz.Check(err)
}

// WriteFile writes the given data to a file at the specified path. It creates
// parent directories if they do not exist. It panics if any error occurs.
func WriteFile(file string, data []byte) error {
	if err := CreateParentDirs(file); err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

// WriteFromReader writes the content from the given `io.Reader` to a file at
// the specified path. It panics if any error occurs.
func WriteFromReader(path string, reader io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	buf := make([]byte, 8*1024)
	_, err = io.CopyBuffer(file, reader, buf)
	return err
}

// Move moves a file from `fileFrom` to `fileTo`. It creates parent directories
// for the destination if they do not exist. It panics if any error occurs.
func Move(fileFrom string, fileTo string) error {
	if err := CreateParentDirs(fileTo); err != nil {
		return err
	}

	in, err := os.Open(fileFrom)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(fileTo)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return Remove(fileFrom)
}

// -----------------------------------------------------------------------------
// Reading
// -----------------------------------------------------------------------------

// ReadAllString reads all content from the given `io.Reader` and returns it as a
// string. It has a default limit of 10MB.
func ReadAllString(r io.Reader) string {
	return ReadAllStringLimited(r, 10*1024*1024)
}

// ReadAllStringLimited reads content from the given `io.Reader` up to a specified
// maximum number of bytes and returns it as a string.
func ReadAllStringLimited(r io.Reader, max int) string {
	var b bytes.Buffer
	b.ReadFrom(io.LimitReader(r, int64(max)))
	return b.String()
}

// ReadAllLimited reads all content from the given `io.Reader` up to a specified
// maximum number of bytes. It panics if the content size exceeds the limit.
func ReadAllLimited(r io.Reader, max int) blobz.Blob {
	body, err := io.ReadAll(io.LimitReader(r, int64(max+1)))
	errorz.Check(err)
	if len(body) > max {
		panic(fmt.Errorf("It is too large. Expected: %d, but was: %d", max, len(body)))
	}
	return blobz.New(body)
}

// TailFile reads the last `size` bytes of a file and returns them as a byte slice.
// If the file is smaller than `size`, it returns the entire file content.
func TailFile(file string, size int) []byte {
	f, err := os.Open(file)
	errorz.Check(err)
	defer f.Close()

	stat, err := f.Stat()
	errorz.Check(err)

	fileSize := stat.Size()
	if int64(size) > fileSize {
		size = int(fileSize)
	}

	if size == 0 {
		return []byte{}
	}

	buf := make([]byte, size)
	offset := fileSize - int64(size)
	_, err = f.ReadAt(buf, offset)
	errorz.Check(err)
	return buf
}

// MustReadFile reads the content of a file up to a specified maximum number of bytes
// and returns it as a `blobz.Blob`. It panics if any error occurs during reading
// or if the file size exceeds the limit.
func MustReadFile(file string, max int) blobz.Blob {
	ret, err := ReadFile(file, max)
	errorz.Check(err)
	return ret
}

func ReadFile(file string, max int) (blobz.Blob, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadAllLimited(f, max), err
}
