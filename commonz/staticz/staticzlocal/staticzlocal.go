package staticzlocal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzentry"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

var ErrNotFound = errors.New("not found")

func LookupCurrentDir(name any) (*pathz.Path, error) {
	original, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}
	current := original
	for {
		p := filepath.Join(current, "stzfiles", "stzfiles.txt")
		content, err := filez.ReadFile(p, 256)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				ncurrent := filepath.Dir(current)
				if ncurrent == "" || ncurrent == current {
					return nil, fmt.Errorf("%w: not found %s", ErrNotFound, original)
				}
				current = ncurrent
				continue
			}
			return nil, err
		}
		strContent := strings.TrimSpace(content.String())
		if strContent != fmt.Sprintf("%s", name) {
			return nil, fmt.Errorf("wrong name, expected: %s, but was: %s", name, strContent)
		}
		ret, err := pathz.Parse(filepath.Join(current, "stzfiles"))
		return ret, err
	}
}

func Walk(ctx context.Context, name any, callback func(entry staticzentry.Entry) error) error {
	dirPath, err := LookupCurrentDir(name)
	if err != nil {
		return err
	}
	dir := dirPath.String()
	return filez.WalkLoader(dir, func(entry filez.WalkLoaderEntry) error {
		if entry.IsDir() {
			return nil
		}
		p, err := filepath.Rel(dir, entry.Path())
		if err != nil {
			return fmt.Errorf("error extracting relative path: %s (%s)", entry.Path(), dir)
		}
		pz, err := pathz.Parse(p)
		if err != nil {
			return fmt.Errorf("error parsing path: %w", err)
		}
		nEntry := newEntry(ctx, name, dirPath, pz, entry.Size())
		return callback(nEntry)
	})
}

func Lookup(ctx context.Context, name any, p *pathz.Path) (staticzentry.Entry, error) {
	err := p.Validate(pathz.ValidateOptions{
		MaxParents:  new(0),
		EndingSlash: new(false),
		Empty:       new(false),
	})
	if err != nil {
		return nil, err
	}
	dir, err := LookupCurrentDir(name)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	np, err := dir.Join(p)
	if err != nil {
		return nil, err
	}
	pstr := np.String()
	fileInfo, err := os.Stat(pstr)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: error looking for file: %s", err, p)
	}
	ret := newEntry(ctx, name, dir, p, fileInfo.Size())
	return ret, nil
}

func newEntry(ctx context.Context, packName any, base *pathz.Path, name *pathz.Path, size int64) staticzentry.Entry {
	fullPath := base.MustJoin(name).String()
	return staticzentry.NewEntry(name, size, func() (io.ReadCloser, error) {
		localData, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, err
		}
		codeEntry, err := staticzloader.Lookup(ctx, packName, name)
		if err != nil {
			return nil, err
		}
		if codeEntry == nil {
			logger.Info(ctx, "static file found locally only", map[string]any{"f": name.String()})
			return io.NopCloser(bytes.NewBuffer(localData)), nil
		}
		codeReader, err := codeEntry.Open()
		if err != nil {
			return nil, err
		}
		codeData, err := io.ReadAll(codeReader)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(localData, codeData) {
			logger.Info(ctx, "static file does not match", map[string]any{"f": name.String()})
		}
		return io.NopCloser(bytes.NewBuffer(localData)), nil
	})
}
