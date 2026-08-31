package deferz

import (
	"context"
	"io"
	"sync"
)

type Entry struct {
	Fn func()
}

type Deferz struct {
	ctx     context.Context
	entries []*Entry
	lock    sync.Mutex
}

func (me *Deferz) AddCloser(closer io.Closer, result func(err error)) {
	me.Add(func() {
		err := closer.Close()
		if result != nil {
			result(err)
		}
	})
}

func (me *Deferz) AddCloserS(closer io.Closer) {
	me.AddCloser(closer, nil)
}

func (me *Deferz) Add(fn func()) {
	me.lock.Lock()
	defer me.lock.Unlock()
	me.entries = append(me.entries, &Entry{Fn: fn})
}

func (me *Deferz) Clean() {
	me.lock.Lock()
	defer me.lock.Unlock()
	me.entries = []*Entry{}
}

func (me *Deferz) Detach() *Deferz {
	me.lock.Lock()
	defer me.lock.Unlock()
	ret := New(me.ctx)
	ret.entries = me.entries
	me.entries = []*Entry{}
	return ret
}

func (me *Deferz) Do() {
	me.lock.Lock()
	defer me.lock.Unlock()
	for _, et := range me.entries {
		defer func() { et.Fn() }()
	}
	me.entries = []*Entry{}
}

func (me *Deferz) Close() error {
	me.Do()
	return nil
}

func New(ctx context.Context) *Deferz {
	return &Deferz{
		ctx:     ctx,
		entries: []*Entry{},
	}
}
