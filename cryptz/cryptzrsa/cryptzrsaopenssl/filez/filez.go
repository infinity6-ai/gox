package filez

import (
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"go.code.infinity6.ai/platform/util/strconvz"
)

func Write(dest string, payload []byte) {
	fd := -1
	if strings.HasPrefix(dest, "@") {
		fdStr := strings.TrimPrefix(dest, "@")
		fd = strconvz.ParseInt(fdStr)
	} else {
		file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		errorz.Check(err)
		defer file.Close()
		fd = int(file.Fd())
	}
	_, err := syscall.Write(fd, payload)
	errorz.Check(err)

}

func CreateTempFile(name string, content []byte) string {
	p := fmt.Sprintf("i6go-platform-tmp-%s-*", name)
	file, err := os.CreateTemp("", p)
	errorz.Check(err)
	if len(content) > 0 {
		Write(file.Name(), content)
	}
	return file.Name()
}

func ReadFile(file string, max int) []byte {
	f, err := os.Open(file)
	errorz.Check(err)
	defer f.Close()
	return ReadAllLimited(f, max)
}

func ReadAllLimited(r io.Reader, max int) []byte {
	body, err := io.ReadAll(io.LimitReader(r, int64(max+1)))
	errorz.Check(err)
	if len(body) > max {
		panic(fmt.Errorf("It is too large. Expected: %d, but was: %d", max, len(body)))
	}
	return body
}
