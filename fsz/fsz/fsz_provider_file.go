package fsz

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
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

	return &FileStat{
		Url:       url,
		Size:      uint64(info.Size()),
		CreatedAt: filez.GetCreatedAt(p),
		UpdatedAt: filez.GetUpdatedAt(p),
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

func (ff *fileFs) Download(ctx context.Context, url *urlz.Url, callback func(found bool, headers http.Header, reader io.Reader) error) error {
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
	defer file.Close()

	headers := make(http.Header)

	return callback(true, headers, file)
}

func (ff *fileFs) Delete(ctx context.Context, url *urlz.Url) error {
	p := url.Path.String()
	return filez.Remove(p)
}

func (ff *fileFs) Ls(ctx context.Context, prefix *urlz.Url) (Paginator, error) {
	panic("implement it")
}

func (ff *fileFs) Find(ctx context.Context, prefix *urlz.Url) (Paginator, error) {
	panic("implement it")
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

func (ff *fileFs) RmTree(ctx context.Context, url *urlz.Url) error {
	p := url.Path.String()
	return os.RemoveAll(p)
}

func (ff *fileFs) Move(ctx context.Context, src *urlz.Url, dest *urlz.Url) error {
	srcPath := src.Path.String()
	destPath := dest.Path.String()

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat source %s: %w", srcPath, err)
	}

	if srcInfo.IsDir() {
		return fmt.Errorf("cannot move directory %s: Move is only for files", srcPath)
	}

	if err := filez.CreateParentDirs(destPath); err != nil {
		return fmt.Errorf("failed to create parent directories for destination %s: %w", destPath, err)
	}

	return os.Rename(srcPath, destPath)
}
