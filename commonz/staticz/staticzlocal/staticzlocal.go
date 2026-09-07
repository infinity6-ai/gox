package staticzlocal

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzentry"
)

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
		nEntry := staticzentry.NewEntry(pz, entry.Size(), entry.Open)
		return callback(nEntry)
	})
}

// func Lookup(ctx context.Context, name any, p *pathz.Path) (staticzentry.Entry, error) {
// 	dir, err := LookupCurrentDir(name)
// 	if err != nil {
// 		if errors.Is(err, ErrNotFound) {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	p = filepath.Join(dir, p)
// 	fileInfo, err := os.Stat(p)
// 	if err != nil {
// 		return nil, fmt.Errorf("%w: error looking for file: %s", err, p)
// 	}
// 	staticzentry.NewEntry()

// }
