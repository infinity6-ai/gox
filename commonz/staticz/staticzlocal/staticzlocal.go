package staticzlocal

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/infinity6-ai/gox/commonz/filez"
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
		return filepath.Join(current, "stzfiles"), nil
	}
}
