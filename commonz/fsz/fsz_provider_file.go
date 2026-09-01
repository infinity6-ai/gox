package fsz

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/urlz"
)

func providerFile() FsProvider {
	return &fileFs{}
}

type fileFs struct{}

func (ff *fileFs) Stat(ctx context.Context, url *urlz.Url) *FileStat {
	p := url.Path.String()
	info := filez.Stat(p)
	if info == nil {
		return nil
	}

	// Read file for content type and MD5
	file, err := os.Open(p)
	errorz.Check(err)
	defer file.Close()

	// content type
	contentTypeBuffer := make([]byte, 512)
	n, err := file.Read(contentTypeBuffer)
	if err != nil && err != io.EOF {
		errorz.Check(err)
	}
	contentType := http.DetectContentType(contentTypeBuffer[:n])

	// MD5
	_, err = file.Seek(0, io.SeekStart)
	errorz.Check(err)
	hash := md5.New()
	_, err = io.Copy(hash, file)
	errorz.Check(err)
	md5sum := hex.EncodeToString(hash.Sum(nil))

	return &FileStat{
		Url:         *url,
		ContentType: contentType,
		Md5:         md5sum,
		Size:        uint64(info.Size()),
		Etag:        md5sum,
		CreatedAt:   filez.GetCreatedAt(p),
		UpdatedAt:   filez.GetUpdatedAt(p),
	}
}

func (ff *fileFs) Upload(ctx context.Context, url *urlz.Url, reader io.Reader) error {
	p := url.Path.String()
	filez.CreateParentDirs(p)
	filez.WriteFromReader(p, reader)
	return nil
}

func (ff *fileFs) Download(ctx context.Context, url *urlz.Url, callback func(found bool, headers http.Header, reader io.Reader)) {
	p := url.Path.String()
	info := filez.Stat(p)
	if info == nil {
		callback(false, nil, nil)
		return
	}

	file, err := os.Open(p)
	errorz.Check(err)

	headers := make(http.Header)
	headers.Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	if fStat := ff.Stat(ctx, url); fStat != nil {
		headers.Set("Content-Type", fStat.ContentType)
		headers.Set("Etag", fStat.Etag)
	}

	callback(true, headers, file)
}

func (ff *fileFs) Delete(ctx context.Context, url *urlz.Url) error {
	p := url.Path.String()
	filez.Remove(p)
	return nil
}

type filePaginator struct {
	files []*FileStat
	pos   int
}

func (p *filePaginator) Paginate(ctx context.Context, max int) []*FileStat {
	if p.pos >= len(p.files) {
		return nil
	}
	end := p.pos + max
	if end > len(p.files) {
		end = len(p.files)
	}
	results := p.files[p.pos:end]
	p.pos = end
	return results
}

func (ff *fileFs) Ls(ctx context.Context, prefix *urlz.Url) Paginator {
	paginator := &filePaginator{
		files: make([]*FileStat, 0),
	}

	dirPath := prefix.Path.String()
	filez.Ls(dirPath, func(idx int, path string, f os.DirEntry) bool {
		info, err := f.Info()
		errorz.Check(err)

		u, err := urlz.Parse("file://" + filepath.ToSlash(path))
		errorz.Check(err)

		fileStat := &FileStat{
			Url:       *u,
			Size:      uint64(info.Size()),
			UpdatedAt: filez.GetUpdatedAt(path),
			CreatedAt: filez.GetCreatedAt(path),
		}
		paginator.files = append(paginator.files, fileStat)
		return false
	})

	return paginator
}

func (ff *fileFs) SignGet(ctx context.Context, url *urlz.Url, duration time.Duration) (string, error) {
	return "", ErrUnsupportedOperation
}

func (ff *fileFs) SignPut(ctx context.Context, url *urlz.Url, duration time.Duration) (string, error) {
	return "", ErrUnsupportedOperation
}

func (ff *fileFs) SignDelete(ctx context.Context, url *urlz.Url, duration time.Duration) (string, error) {
	return "", ErrUnsupportedOperation
}
