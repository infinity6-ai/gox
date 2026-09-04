package fsz

import (
	"context"
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

// Paginator for Ls (non-recursive, streaming)
type fileLsPaginator struct {
	dir     *os.File
	dirPath string
	opened  bool
}

func (p *fileLsPaginator) Close() error {
	if p.dir != nil {
		return p.dir.Close()
	}
	return nil
}

func (p *fileLsPaginator) Paginate(ctx context.Context, max int) ([]*FileStat, error) {
	if !p.opened {
		dir, err := os.Open(p.dirPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil // Empty page for non-existent dir
			}
			return nil, fmt.Errorf("failed to open directory %s: %w", p.dirPath, err)
		}
		p.dir = dir
		p.opened = true
	}

	if p.dir == nil {
		return nil, nil
	}

	entries, err := p.dir.Readdir(max)
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var results []*FileStat
	for _, info := range entries {
		path := filepath.Join(p.dirPath, info.Name())
		u, err := urlz.Parse("file://" + filepath.ToSlash(path))
		if err != nil {
			continue
		}
		results = append(results, &FileStat{
			Url:       u,
			Size:      uint64(info.Size()),
			UpdatedAt: filez.GetUpdatedAt(path),
			CreatedAt: filez.GetCreatedAt(path),
		})
	}
	return results, nil
}

func (ff *fileFs) Ls(ctx context.Context, prefix *urlz.Url) (Paginator, error) {
	dirPath := prefix.Path.String()
	return &fileLsPaginator{dirPath: dirPath}, nil
}

type filePaginatorItem struct {
	stat *FileStat
	err  error
}

// Paginator for Find (recursive, streaming)
type fileFindPaginator struct {
	ch     <-chan *filePaginatorItem
	cancel context.CancelFunc
}

func (p *fileFindPaginator) Close() error {
	p.cancel()
	return nil
}

func (p *fileFindPaginator) Paginate(ctx context.Context, max int) ([]*FileStat, error) {
	var results []*FileStat
	for i := 0; i < max; i++ {
		select {
		case item, ok := <-p.ch:
			if !ok {
				return results, nil
			}
			if item.err != nil {
				return nil, item.err
			}
			results = append(results, item.stat)
		case <-ctx.Done():
			return results, ctx.Err()
		}
	}
	return results, nil
}

func (ff *fileFs) Find(ctx context.Context, prefix *urlz.Url) (Paginator, error) {
	ch := make(chan *filePaginatorItem)
	pCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer close(ch)
		dirPath := prefix.Path.String()
		err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return fmt.Errorf("failed to get file info for %s: %w", path, err)
			}

			u, err := urlz.Parse("file://" + filepath.ToSlash(path))
			if err != nil {
				return fmt.Errorf("failed to parse url for %s: %w", path, err)
			}

			select {
			case ch <- &filePaginatorItem{stat: &FileStat{
				Url:       u,
				Size:      uint64(info.Size()),
				UpdatedAt: filez.GetUpdatedAt(path),
				CreatedAt: filez.GetCreatedAt(path),
			}}:
			case <-pCtx.Done():
				return pCtx.Err()
			}
			return nil
		})
		if err != nil {
			if os.IsNotExist(err) {
				return // Gracefully exit, resulting in an empty paginator
			}
			ch <- &filePaginatorItem{err: fmt.Errorf("failed to walk directory %s: %w", dirPath, err)}
		}
	}()

	return &fileFindPaginator{ch: ch, cancel: cancel}, nil
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
