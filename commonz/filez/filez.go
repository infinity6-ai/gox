package filez

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"go.code.infinity6.ai/platform/util/strconvz"
	"golang.org/x/sys/unix"
)

func Remove(path string) bool {
	err := os.Remove(path)
	if err != nil {
		e, ok := err.(*os.PathError)
		if ok && e.Err == syscall.ENOENT {
			return false
		} else if ok && e.Err == syscall.ENOTDIR {
			return false
		}
		errorz.Check(err)
	}
	return true
}

func Parent(p string) string {
	return filepath.Dir(p)
}

func CreateParentDirs(file string) {
	dir := Parent(file)
	errorz.Check(os.MkdirAll(dir, os.ModePerm))
}

func FileExists(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false
		}
		errorz.Check(err)
	}
	return true
}

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

func RmTree(dir string) error {
	return os.RemoveAll(dir)
}

func FindParent(filename string, startDir string) (string, bool) {
	dir := startDir
	if !path.IsAbs(dir) {
		d, err := filepath.Abs(dir)
		errorz.Check(err)
		dir = d
	}
	for {
		fullpath := filepath.Join(dir, filename)
		if FileExists(fullpath) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func Walk(base string, callback func(path string, f fs.DirEntry) bool) {
	filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if !os.IsNotExist(err) {
				errorz.Check(err)
			}
			return nil
		}
		stop := callback(path, d)
		if stop {
			return filepath.SkipDir
		}
		return nil
	})
}

func Ls(dir string, callback func(idx int, path string, f fs.DirEntry) bool) {
	files, err := os.ReadDir(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			errorz.Check(err)
		}
	}
	for idx, file := range files {
		path := filepath.Join(dir, file.Name())
		if callback(idx, path, file) {
			return
		}
	}
}

func DirList(dir string, regex string) []string {
	ret := []string{}
	files, err := os.ReadDir(dir)
	errorz.Check(err)
	for _, file := range files {
		name := file.Name()
		match, err := regexp.MatchString(regex, name)
		errorz.Check(err)
		if match {
			ret = append(ret, file.Name())
		}
	}
	return ret
}

func DirListLimited(dir string, regex string, limit int) []string {
	ret := []string{}
	files, err := os.ReadDir(dir)
	errorz.Check(err)
	for _, file := range files {
		name := file.Name()
		match, err := regexp.MatchString(regex, name)
		errorz.Check(err)
		if match {
			ret = append(ret, file.Name())
			if len(ret) >= limit {
				return ret
			}
		}
	}
	return ret
}

func CreateTempDir(name string) string {
	p := fmt.Sprintf("gox-tmp-%s-*", name)
	dir, err := os.MkdirTemp("", p)
	errorz.Check(err)
	return dir
}

func CreateTempFile(name string, content []byte) string {
	p := fmt.Sprintf("gox-tmp-%s-*", name)
	file, err := os.CreateTemp("", p)
	errorz.Check(err)
	if len(content) > 0 {
		Write(file.Name(), content)
	}
	return file.Name()
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		errorz.Check(err)
	}
	return info.IsDir()
}

func Stat(path string) fs.FileInfo {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		errorz.Check(err)
	}
	return info
}

func IsDirEmpty(path string) bool {
	dir, err := os.Open(path)
	errorz.Check(err)
	defer dir.Close()

	contents, err := dir.Readdirnames(1)
	if err != nil {
		if err == io.EOF {
			return true
		}
		errorz.Check(err)
	}

	return len(contents) == 0
}

func ReadAllString(r io.Reader) string {
	return ReadAllStringLimited(r, 10*1024*1024)
}

func ReadAllStringLimited(r io.Reader, max int) string {
	var b bytes.Buffer
	b.ReadFrom(io.LimitReader(r, int64(max)))
	return b.String()
}

func WriteFile(file string, data []byte) {
	CreateParentDirs(file)
	err := os.WriteFile(file, data, os.ModePerm)
	errorz.Check(err)
}

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

func ReadAllLimited(r io.Reader, max int) []byte {
	body, err := io.ReadAll(io.LimitReader(r, int64(max+1)))
	errorz.Check(err)
	if len(body) > max {
		panic(fmt.Errorf("It is too large. Expected: %d, but was: %d", max, len(body)))
	}
	return body
}

func Move(fileFrom string, fileTo string) {
	CreateParentDirs(fileTo)

	in, err := os.Open(fileFrom)
	errorz.Check(err)
	defer in.Close()

	out, err := os.Create(fileTo)
	errorz.Check(err)
	defer out.Close()

	_, err = io.Copy(out, in)
	errorz.Check(err)

	Remove(fileFrom)
}

func ReadFile(file string, max int) blobz.Blob {
	f, err := os.Open(file)
	errorz.Check(err)
	defer f.Close()
	return blobz.New(ReadAllLimited(f, max))
}

func Size(filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return -1
		}
		errorz.Check(err)
	}
	return info.Size()
}

func WriteFromReader(path string, reader io.Reader) {
	file, err := os.Create(path)
	errorz.Check(err)
	defer file.Close()
	buf := make([]byte, 8*1024)
	_, err = io.CopyBuffer(file, reader, buf)
	errorz.Check(err)
}

func CreateFileFromPath(path string) *os.File {
	CreateParentDirs(path)

	file, err := os.Create(path)
	errorz.Check(err)
	return file
}

func GetCreatedAt(path string) *time.Time {
	var stat unix.Statx_t
	err := unix.Statx(unix.AT_FDCWD, path, unix.AT_STATX_SYNC_AS_STAT, unix.STATX_BTIME, &stat)
	errorz.Check(err)
	if stat.Mask&unix.STATX_BTIME == 0 {
		// not supported
		return nil
	}
	ret := time.Unix(stat.Btime.Sec, int64(stat.Btime.Nsec))
	return &ret
}

func GetUpdatedAt(path string) *time.Time {
	fi, err := os.Stat(path)
	errorz.Check(err)
	updatedTime := fi.ModTime()
	return &updatedTime
}
