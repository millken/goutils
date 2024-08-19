package accumulator

import (
	"sync"
	"time"
)

const bucketsCount = 512

type entry struct {
	previous     *window
	current      *window
	limit        uint64
	windowLength time.Duration
}

func newEntry(limit uint64, windowLength time.Duration) *entry {
	zero := time.Unix(0, 0)
	return &entry{
		previous:     newWindow(0, zero),
		current:      newWindow(0, zero),
		limit:        limit,
		windowLength: windowLength,
	}
}

type bucket struct {
	mu sync.RWMutex
	m  map[uint64]*entry
}

func (b *bucket) Reset() {
	b.mu.Lock()
	b.m = make(map[uint64]*entry)
	b.mu.Unlock()
}

func (b *bucket) Get(key uint64) (*entry, bool) {
	b.mu.RLock()
	e, ok := b.m[key]
	b.mu.RUnlock()
	return e, ok
}

func (b *bucket) Set(key uint64, e *entry) {
	b.mu.Lock()
	b.m[key] = e
	b.mu.Unlock()
}
