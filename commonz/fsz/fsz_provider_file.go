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

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/urlz"
)

func providerFile() FsProvider {
	return &fileFs{}
}

type fileFs struct{}

func (ff *fileFs) Stat(ctx context.Context, url *urlz.Url) (*FileStat, error) {
	p := url.Path.String()
	info := filez.Stat(p)
	if info == nil {
		return nil, nil
	}

	// Read file for content type and MD5
	file, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", p, err)
	}
	defer file.Close()

	// content type
	contentTypeBuffer := make([]byte, 512)
	n, err := file.Read(contentTypeBuffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file for content type detection %s: %w", p, err)
	}
	contentType := http.DetectContentType(contentTypeBuffer[:n])

	// MD5
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek file %s: %w", p, err)
	}
	hash := md5.New()
	if _, err = io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("failed to copy file for md5 calculation %s: %w", p, err)
	}
	md5sum := hex.EncodeToString(hash.Sum(nil))

	return &FileStat{
		Url:         *url,
		ContentType: contentType,
		Md5:         md5sum,
		Size:        uint64(info.Size()),
		Etag:        md5sum,
		CreatedAt:   filez.GetCreatedAt(p),
		UpdatedAt:   filez.GetUpdatedAt(p),
	}, nil
}

func (ff *fileFs) Upload(ctx context.Context, url *urlz.Url, headers http.Header, reader io.Reader) error {
	p := url.Path.String()
	if err := filez.CreateParentDirs(p); err != nil {
		return fmt.Errorf("failed to create parent directories for %s: %w", p, err)
	}
	if err := filez.WriteFromReader(p, reader); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", p, err)
	}
	return nil
}

func (ff *fileFs) Download(ctx context.Context, url *urlz.Url, callback func(found bool, headers http.Header, reader io.Reader)) error {
	p := url.Path.String()
	info := filez.Stat(p)
	if info == nil {
		callback(false, nil, nil)
		return nil
	}

	file, err := os.Open(p)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", p, err)
	}

	headers := make(http.Header)
	headers.Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	if fStat, err := ff.Stat(ctx, url); err == nil && fStat != nil {
		headers.Set("Content-Type", fStat.ContentType)
		headers.Set("Etag", fStat.Etag)
	}

	callback(true, headers, file)
	return nil
}

func (ff *fileFs) Delete(ctx context.Context, url *urlz.Url) error {
	p := url.Path.String()
	return filez.Remove(p)
}

type filePaginator struct {
	files []*FileStat
	pos   int
}

func (p *filePaginator) Paginate(ctx context.Context, max int) ([]*FileStat, error) {
	if p.pos >= len(p.files) {
		return nil, nil
	}
	end := p.pos + max
	if end > len(p.files) {
		end = len(p.files)
	}
	results := p.files[p.pos:end]
	p.pos = end
	return results, nil
}

func (ff *fileFs) Ls(ctx context.Context, prefix *urlz.Url) (Paginator, error) {
	paginator := &filePaginator{
		files: make([]*FileStat, 0),
	}

	dirPath := prefix.Path.String()
	err := filez.Ls(dirPath, func(idx int, path string, f os.DirEntry) (bool, error) {
		info, err := f.Info()
		if err != nil {
			return false, fmt.Errorf("failed to get file info for %s: %w", path, err)
		}

		u, err := urlz.Parse("file://" + filepath.ToSlash(path))
		if err != nil {
			return false, fmt.Errorf("failed to parse url for %s: %w", path, err)
		}

		fileStat := &FileStat{
			Url:       *u,
			Size:      uint64(info.Size()),
			UpdatedAt: filez.GetUpdatedAt(path),
			CreatedAt: filez.GetCreatedAt(path),
		}
		paginator.files = append(paginator.files, fileStat)
		return false, nil
	})

	if err != nil {
		return nil, err
	}

	return paginator, nil
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

func (ff *fileFs) Copy(ctx context.Context, src *urlz.Url, dest *urlz.Url) error {
	srcPath := src.Path.String()
	destPath := dest.Path.String()

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	if err := filez.CreateParentDirs(destPath); err != nil {
		return fmt.Errorf("failed to create parent directories for destination %s: %w", destPath, err)
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy content from %s to %s: %w", srcPath, destPath, err)
	}

	return nil
}
