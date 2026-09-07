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

func LookupCurrentDir(name any) (string, error) {
	original, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	current := original
	for {
		p := filepath.Join(current, "stzfiles", "stzfiles.txt")
		content, err := filez.ReadFile(p, 256)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				ncurrent := filepath.Dir(current)
				if ncurrent == "" || ncurrent == current {
					return "", fmt.Errorf("%w: not found %s", err, original)
				}
				current = ncurrent
				continue
			}
			return "", err
		}
		strContent := strings.TrimSpace(content.String())
		if strContent != fmt.Sprintf("%s", name) {
			return "", fmt.Errorf("wrong name, expected: %s, but was: %s", name, strContent)
		}
		return filepath.Clean(filepath.Join(current, "stzfiles")), nil
	}
}

func Walk(ctx context.Context, name any, callback func(entry staticzentry.Entry) error) error {
	dir, err := LookupCurrentDir(name)
	if err != nil {
		return err
	}
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
