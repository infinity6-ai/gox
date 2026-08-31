package logzlast

import (
	"sync"

	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
)

type LastEntries struct {
	entries []*logzspec.Entry
	mu      sync.Mutex
}

func (l *LastEntries) Entries() []*logzspec.Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	ret := l.entries
	l.entries = make([]*logzspec.Entry, 0, len(ret))
	return ret
}

func (l *LastEntries) Add(entry *logzspec.Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) >= 5 {
		l.entries = l.entries[1:]
	}
	l.entries = append(l.entries, entry)
}
