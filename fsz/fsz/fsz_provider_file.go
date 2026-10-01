package fsz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// Lister implements real performant directory listing
func (ff *fileFs) Lister(ctx context.Context, prefix *urlz.Url, walker Walker) error {
	if walker.Pager == nil {
		return nil
	}
	pageSize := walker.PageSize
	if pageSize <= 0 {
		pageSize = 1000
	}

	dirPath := filepath.Clean(prefix.Path.String())
	cursor := cleanCursor(dirPath, walker.StartCursor)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var currentPage []*FileStat

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context error during lister: %w", err)
		}

		name := entry.Name()
		if cursor != "" && name <= cursor {
			continue
		}

		if len(currentPage) == pageSize {
			lastPath := currentPage[len(currentPage)-1].Url.Path.String()
			rel, err := filepath.Rel(dirPath, lastPath)
			var nextCursor string
			if err == nil {
				nextCursor = filepath.ToSlash(rel)
			} else {
				nextCursor = filepath.ToSlash(lastPath)
			}

			err = walker.Pager(currentPage, nextCursor)
			if err != nil {
				if errors.Is(err, ErrStop) {
					return nil
				}
				return err
			}
			currentPage = make([]*FileStat, 0, pageSize)
		}

		path := filepath.Join(dirPath, name)
		info, err := entry.Info()
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to get file info for %s: %w", path, err)
		}

		urlStr := "file://" + filepath.ToSlash(path)
		if info.IsDir() {
			urlStr += "/"
		}

		u, err := urlz.Parse(urlStr)
		if err != nil {
			return fmt.Errorf("failed to parse url for %s: %w", path, err)
		}

		stat := &FileStat{
			Url:       u,
			Size:      uint64(info.Size()),
			UpdatedAt: filez.GetUpdatedAt(path),
			CreatedAt: filez.GetCreatedAt(path),
		}
		if info.IsDir() {
			stat.Size = 0
		}

		currentPage = append(currentPage, stat)
	}

	if len(currentPage) > 0 {
		err = walker.Pager(currentPage, "")
		if err != nil && !errors.Is(err, ErrStop) {
			return err
		}
	}

	return nil
}

// Finder implements real performant recursive directory walking
func (ff *fileFs) Finder(ctx context.Context, prefix *urlz.Url, walker Walker) error {
	if walker.Pager == nil {
		return nil
	}
	pageSize := walker.PageSize
	if pageSize <= 0 {
		pageSize = 1000
	}

	dirPath := filepath.Clean(prefix.Path.String())
	cursor := cleanCursor(dirPath, walker.StartCursor)

	var currentPage []*FileStat
	stopped := false

	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context error during finder: %w", err)
		}

		if path == dirPath {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}
		relSlash := filepath.ToSlash(relPath)

		if d.IsDir() {
			if shouldSkipDir(relSlash, cursor) {
				return filepath.SkipDir
			}
			return nil
		}

		if cursor != "" && compareWalkOrder(relSlash, cursor) <= 0 {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("failed to get file info for %s: %w", path, err)
		}

		if info.IsDir() {
			return nil
		}

		if len(currentPage) == pageSize {
			lastPath := currentPage[len(currentPage)-1].Url.Path.String()
			rel, err := filepath.Rel(dirPath, lastPath)
			var nextCursor string
			if err == nil {
				nextCursor = filepath.ToSlash(rel)
			} else {
				nextCursor = filepath.ToSlash(lastPath)
			}

			pageToSend := currentPage
			currentPage = make([]*FileStat, 0, pageSize)

			pagerErr := walker.Pager(pageToSend, nextCursor)
			if pagerErr != nil {
				if errors.Is(pagerErr, ErrStop) {
					stopped = true
					return filepath.SkipAll
				}
				return pagerErr
			}
		}

		u, err := urlz.Parse("file://" + filepath.ToSlash(path))
		if err != nil {
			return fmt.Errorf("failed to parse url for %s: %w", path, err)
		}

		currentPage = append(currentPage, &FileStat{
			Url:       u,
			Size:      uint64(info.Size()),
			UpdatedAt: filez.GetUpdatedAt(path),
			CreatedAt: filez.GetCreatedAt(path),
		})

		return nil
	})

	if err != nil {
		if os.IsNotExist(err) || errors.Is(err, ErrStop) {
			return nil
		}
		return fmt.Errorf("failed to walk directory %s: %w", dirPath, err)
	}

	if !stopped && len(currentPage) > 0 {
		pagerErr := walker.Pager(currentPage, "")
		if pagerErr != nil && !errors.Is(pagerErr, ErrStop) {
			return pagerErr
		}
	}

	return nil
}

func cleanCursor(baseDir string, cursor string) string {
	if cursor == "" {
		return ""
	}
	if strings.HasPrefix(cursor, "file://") {
		if u, err := urlz.Parse(cursor); err == nil {
			cursor = u.Path.String()
		} else {
			cursor = strings.TrimPrefix(cursor, "file://")
		}
	}
	cleanBase := filepath.Clean(baseDir)
	if filepath.IsAbs(cursor) {
		if rel, err := filepath.Rel(cleanBase, filepath.Clean(cursor)); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
	}
	return strings.Trim(filepath.ToSlash(filepath.Clean(cursor)), "/")
}

func compareWalkOrder(pathA, pathB string) int {
	cleanA := strings.Trim(filepath.ToSlash(pathA), "/")
	cleanB := strings.Trim(filepath.ToSlash(pathB), "/")
	if cleanA == cleanB {
		return 0
	}
	if cleanA == "" {
		return -1
	}
	if cleanB == "" {
		return 1
	}
	partsA := strings.Split(cleanA, "/")
	partsB := strings.Split(cleanB, "/")
	minLen := len(partsA)
	if len(partsB) < minLen {
		minLen = len(partsB)
	}
	for i := 0; i < minLen; i++ {
		if partsA[i] != partsB[i] {
			if partsA[i] < partsB[i] {
				return -1
			}
			return 1
		}
	}
	if len(partsA) < len(partsB) {
		return -1
	} else if len(partsA) > len(partsB) {
		return 1
	}
	return 0
}

func shouldSkipDir(dirRelPath string, cursor string) bool {
	if cursor == "" {
		return false
	}
	dirClean := strings.Trim(filepath.ToSlash(dirRelPath), "/")
	cursorClean := strings.Trim(filepath.ToSlash(cursor), "/")
	if dirClean == "" || dirClean == "." || cursorClean == "" {
		return false
	}
	dirParts := strings.Split(dirClean, "/")
	cursorParts := strings.Split(cursorClean, "/")
	minLen := len(dirParts)
	if len(cursorParts) < minLen {
		minLen = len(cursorParts)
	}
	for i := 0; i < minLen; i++ {
		if dirParts[i] != cursorParts[i] {
			if dirParts[i] < cursorParts[i] {
				return true
			}
			return false
		}
	}
	return false
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
