package regexpz

import (
	"fmt"
	"sync"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

type Collection struct {
	coll map[string]*Regexp
	mu   sync.RWMutex
}

func New(buffer int) *Collection {
	return &Collection{
		coll: make(map[string]*Regexp, buffer),
	}
}

var root = New(0)

func (c *Collection) internalGet(pattern string) *Regexp {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.coll[pattern]
}

func (c *Collection) Get(pattern string, re *Regexp) (*Regexp, error) {
	ret := c.internalGet(pattern)
	if ret != nil {
		return ret, nil
	}
	ret, err := NewRegexp(pattern)
	if err != nil {
		return nil, fmt.Errorf("%w: error compiling regexp: %s", err, pattern)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.coll[pattern] = ret
	return ret, nil
}

func (c *Collection) Must(pattern string, re *Regexp) *Regexp {
	ret, err := c.Get(pattern, re)
	errorz.Check(err)
	return ret
}

func Get(pattern string, re *Regexp) (*Regexp, error) {
	return root.Get(pattern, re)
}

func Must(pattern string, re *Regexp) *Regexp {
	return root.Must(pattern, re)
}
