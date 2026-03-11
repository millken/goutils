//go:build faststack_debug

package faststack

import (
	"fmt"
	"os"
	"sync/atomic"
)

var (
	debugCacheHits   atomic.Uint64
	debugCacheMisses atomic.Uint64
	debugEnabled     = os.Getenv("FASTSTACK_DEBUG") != ""
)

func debugCacheHit() {
	if !debugEnabled {
		return
	}
	debugCacheHits.Add(1)
}

func debugCacheMiss() {
	if !debugEnabled {
		return
	}
	debugCacheMisses.Add(1)
}

// DebugStats returns cache statistics in debug mode.
// Enable with: FASTSTACK_DEBUG=1 go build -tags=faststack_debug
func DebugStats() string {
	hits := debugCacheHits.Load()
	misses := debugCacheMisses.Load()
	total := hits + misses
	if total == 0 {
		return "no cache activity"
	}
	return fmt.Sprintf("hits: %d, misses: %d, hit_rate: %.2f%%",
		hits, misses, float64(hits)*100/float64(total))
}
