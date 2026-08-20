package pathz

import (
	"path"
	"strings"
)

// Clean is a convenience function that makes a path safe by cleaning it.
//
// Effectively, it's a wrapper around path.Clean that also removes any leading
// or trailing slashes.
func Clean(p string) string {
	p = path.Clean(p)
	for strings.HasPrefix(p, "../") {
		p = p[3:]
	}
	p = trimSlash(p)
	if p == "" || p == ".." {
		return "."
	}
	return p
}
