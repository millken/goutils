package faststack

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"unsafe"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=Mode

// Mode controls the behavior of faststack at runtime.
// This can be set at init time or changed at runtime (with atomic cost).
type Mode int

const (
	// ModeEnabled is the default mode with caching enabled.
	ModeEnabled Mode = iota
	// ModeDisabled completely disables all stack operations.
	// Calls return zero values immediately.
	ModeDisabled
	// ModeNoCache enables stack operations without caching.
	ModeNoCache
	// ModeDebug enables additional validation (for testing only).
	ModeDebug
)

var (
	// CurrentMode can be changed at runtime.
	// Use atomic.Load/Store for thread safety.
	CurrentMode atomic.Int32 // Mode
)

func init() {
	// Default to enabled
	CurrentMode.Store(int32(ModeEnabled))
}

// SetMode changes the current mode.
func SetMode(m Mode) {
	CurrentMode.Store(int32(m))
}

type (
	// PC is a program counter alias.
	// Function name, file name and line can be obtained from it but only in the same binary where Caller or FuncEntry was called.
	PC uintptr

	// PCs is a stack trace.
	// It's quiet the same as runtime.CallerFrames but more efficient.
	PCs []PC
)

type (
	nfl struct {
		name string
		file string
		line int
	}
)

// ============================================================================
// Cache Implementation: Sharded for reduced lock contention
// ============================================================================

const (
	cacheShards     = 32 // Power of 2 for fast modulo
	cacheShardMask  = cacheShards - 1
	cacheMaxEntries = 128 // Per shard, total ~4K entries
)

// cacheShard is a single shard of the cache with its own lock.
type cacheShard struct {
	// Use array instead of map for better cache locality
	// Open addressing linear probing
	entries [cacheMaxEntries]cacheEntry
	// Use uint32 for atomic access (fits in atomic.Int32 on 32-bit)
	// We use byte array to avoid alignment issues
	_ [cacheMaxEntries - 1]struct{}
}

type cacheEntry struct {
	pc   PC
	nfl  nfl
	used atomic.Uint32 // 0 = empty, 1 = used
}

var cacheShardsP [cacheShards]cacheShard

// getCache retrieves from sharded cache without locks.
func getCache(pc PC) (nfl, bool) {
	if CurrentMode.Load() == int32(ModeNoCache) {
		return nfl{}, false
	}

	// Fast path: shard selection
	shardIdx := uint32(pc>>3) & cacheShardMask
	shard := &cacheShardsP[shardIdx]

	// Linear probe within shard
	start := (uint32(pc) >> 4) % cacheMaxEntries
	for i := range uint32(8) { // Probe at most 8 entries
		idx := (start + i) % cacheMaxEntries
		e := &shard.entries[idx]

		if e.used.Load() == 0 {
			break
		}
		if e.pc == pc {
			return e.nfl, true
		}
	}

	return nfl{}, false
}

// putCache stores into sharded cache without locks.
func putCache(pc PC, v nfl) {
	if CurrentMode.Load() == int32(ModeNoCache) {
		return
	}

	shardIdx := uint32(pc>>3) & cacheShardMask
	shard := &cacheShardsP[shardIdx]

	// Linear probe for empty slot or matching entry
	start := (uint32(pc) >> 4) % cacheMaxEntries
	var firstEmpty *cacheEntry

	for i := range uint32(8) {
		idx := (start + i) % cacheMaxEntries
		e := &shard.entries[idx]

		used := e.used.Load()
		if used == 0 {
			if firstEmpty == nil {
				firstEmpty = e
			}
			continue
		}
		if e.pc == pc {
			// Update existing
			e.nfl = v
			return
		}
	}

	// Insert into first empty slot or replace randomly
	if firstEmpty != nil {
		firstEmpty.pc = pc
		firstEmpty.nfl = v
		firstEmpty.used.Store(1)
	}
}

// ============================================================================
// NameFileLine: Hot path optimized
// ============================================================================

// NameFileLine returns function name, file and line number for location.
//
// This works only in the same binary where location was captured.
//
//go:nosplit
//go:inline
func (l PC) NameFileLine() (name, file string, line int) {
	if l == 0 || CurrentMode.Load() == int32(ModeDisabled) {
		return
	}

	// Fast path: cache hit
	if c, ok := getCache(l); ok {
		return c.name, c.file, c.line
	}

	// Slow path: fetch from runtime
	name, file, line = l.nameFileLineInternal()

	if file != "" {
		file = cropFilename(file, name)
	}

	putCache(l, nfl{name: name, file: file, line: line})
	return
}

// nameFileLineInternal is separated to allow inlining of NameFileLine.
func (l PC) nameFileLineInternal() (name, file string, line int) {
	// Inline the CallersFrames call to avoid allocation
	// We reuse a static buffer (not thread-safe, but ok for single goroutine)
	var pcBuf [1]uintptr
	pcBuf[0] = uintptr(l)

	fs := runtime.CallersFrames(pcBuf[:])
	f, _ := fs.Next()
	return f.Function, f.File, f.Line
}

// nameFileLine is exposed for benchmarks (no cache path).
func (l PC) nameFileLine() (name, file string, line int) {
	return l.nameFileLineInternal()
}

// ============================================================================
// FuncEntry: Zero allocation
// ============================================================================

// FuncEntry returns the entry PC of the function containing pc.
//
//go:nosplit
//go:inline
func (l PC) FuncEntry() PC {
	if l == 0 || CurrentMode.Load() == int32(ModeDisabled) {
		return 0
	}

	f := runtime.FuncForPC(uintptr(l))
	if f == nil {
		return 0
	}
	return PC(f.Entry())
}

// ============================================================================
// Caller: Zero allocation, maximum inlining
// ============================================================================

// Caller returns information about the calling goroutine's stack.
// The argument s is the number of frames to ascend, with 0 identifying the caller of Caller.
//
//go:nosplit
//go:inline
func Caller(s int) (r PC) {
	if CurrentMode.Load() == int32(ModeDisabled) {
		return 0
	}
	caller1(1+s, &r, 1, 1)
	return
}

// FuncEntry returns the entry PC of the caller at skip s.
//
//go:nosplit
func FuncEntry(s int) (r PC) {
	if CurrentMode.Load() == int32(ModeDisabled) {
		return 0
	}
	caller1(1+s, &r, 1, 1)
	return r.FuncEntry()
}

// ============================================================================
// CallerOnce: One-time initialization helpers
// ============================================================================

// CallerOnce stores the caller PC once, atomically.
// Useful for one-time initialization where you want to capture caller once.
//
//go:nosplit
func CallerOnce(s int, pc *PC) (r PC) {
	r = PC(atomic.LoadUintptr((*uintptr)(unsafe.Pointer(pc))))
	if r != 0 {
		return
	}

	r = Caller(1 + s)
	if r == 0 {
		return 0
	}

	// Atomic compare-and-swap
	for {
		old := PC(atomic.LoadUintptr((*uintptr)(unsafe.Pointer(pc))))
		if old != 0 {
			return old
		}
		if atomic.CompareAndSwapUintptr((*uintptr)(unsafe.Pointer(pc)), 0, uintptr(r)) {
			return r
		}
	}
}

// FuncEntryOnce is like CallerOnce but returns function entry.
//
//go:nosplit
func FuncEntryOnce(s int, pc *PC) (r PC) {
	r = PC(atomic.LoadUintptr((*uintptr)(unsafe.Pointer(pc))))
	if r != 0 {
		return r // Already cached
	}

	r = Caller(1 + s)
	if r == 0 {
		return 0
	}
	r = r.FuncEntry()

	// Atomic compare-and-swap
	for {
		old := PC(atomic.LoadUintptr((*uintptr)(unsafe.Pointer(pc))))
		if old != 0 {
			return old
		}
		if atomic.CompareAndSwapUintptr((*uintptr)(unsafe.Pointer(pc)), 0, uintptr(r)) {
			return r
		}
	}
}

// ============================================================================
// Callers: Zero-allocation variants
// ============================================================================

// Callers returns callers stack trace.
//
//go:nosplit
func Callers(skip, n int) PCs {
	if CurrentMode.Load() == int32(ModeDisabled) || n <= 0 {
		return nil
	}

	tr := make(PCs, n)
	n = callers(1+skip, tr)
	return tr[:n]
}

// CallersView returns a view of the caller stack without allocation.
// The returned slice is only valid until the next Call to CallersView.
// Use with caution in concurrent code.
//
//go:nosplit
func CallersView(skip, n int) PCs {
	if CurrentMode.Load() == int32(ModeDisabled) || n <= 0 {
		return nil
	}

	// Use a pre-allocated buffer for the view
	// Not thread-safe, but documented as such
	const maxViewDepth = 32
	var buf [maxViewDepth]PC

	if n > maxViewDepth {
		n = maxViewDepth
	}

	n = callers(1+skip, buf[:n])
	return buf[:n]
}

// CallersFill puts callers stack trace into provided slice.
// This is the only truly zero-allocation method.
//
//go:nosplit
func CallersFill(skip int, tr PCs) PCs {
	if CurrentMode.Load() == int32(ModeDisabled) || len(tr) == 0 {
		return nil
	}
	n := callers(1+skip, tr)
	return tr[:n]
}

// ============================================================================
// cropFilename: Hot path optimization
// ============================================================================

func cropFilename(fn, tp string) string {
	p := strings.LastIndexByte(tp, '/')
	pp := strings.IndexByte(tp[p+1:], '.')
	if pp == -1 {
		// No type information (e.g., "package.function" without receiver type)
		// Keep only the package path including trailing slash
		tp = tp[:p+1]
	} else {
		tp = tp[:p+1+pp] // cut type and func name
	}

	for {
		if p = strings.LastIndex(fn, tp); p != -1 {
			return fn[p:]
		}

		p = strings.IndexByte(tp, '/')
		if p == -1 {
			return filepath.Base(fn)
		}

		tp = tp[p+1:]
	}
}
