package faststack

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

// pcsPool is a pool for reusing PC slices to reduce allocations.
var pcsPool = sync.Pool{
	New: func() any {
		s := make(PCs, 64)
		return &s
	},
}

// framesPool is a pool for reusing uintptr slices in CallersFrames.
var framesPool = sync.Pool{
	New: func() any {
		s := make([]uintptr, 1)
		return &s
	},
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

var (
	// Fast path cache using sync.Map for better read performance
	cache sync.Map // map[PC]*cacheEntry

	// Counter for periodic cache cleanup
	cacheCount atomic.Uint64
	cacheMax  = 4096 // Higher limit for better cache hit rate
)

// cacheEntry stores cached location info with access timestamp.
type cacheEntry struct {
	nfl  nfl
	last atomic.Uint64
}

const cacheCleanupThreshold = 10000

func getCache(pc PC) (nfl, bool) {
	v, ok := cache.Load(pc)
	if !ok {
		return nfl{}, false
	}

	e := v.(*cacheEntry)
	// Update access time (no need to be strictly accurate)
	cnt := cacheCount.Add(1)
	e.last.Store(cnt)

	return e.nfl, true
}

func putCache(pc PC, v nfl) {
	// Periodic cleanup to prevent unbounded growth
	cnt := cacheCount.Add(1)
	if cnt%cacheCleanupThreshold == 0 {
		cleanupCache(cnt - cacheCleanupThreshold)
	}

	e := &cacheEntry{nfl: v}
	e.last.Store(cnt)
	cache.Store(pc, e)
}

// cleanupCache removes entries older than threshold.
func cleanupCache(threshold uint64) {
	// Delete a batch of old entries to amortize cleanup cost
	deleted := 0
	maxToDelete := 256

	cache.Range(func(key, value any) bool {
		if deleted >= maxToDelete {
			return false
		}
		e := value.(*cacheEntry)
		if e.last.Load() < threshold {
			cache.Delete(key)
			deleted++
		}
		return true
	})
}

// NameFileLine returns function name, file and line number for location.
//
// This works only in the same binary where location was captured.
//
// This functions is a little bit modified version of runtime.(*Frames).Next().
func (l PC) NameFileLine() (name, file string, line int) {
	if l == 0 {
		return
	}

	if c, ok := getCache(l); ok {
		return c.name, c.file, c.line
	}

	name, file, line = l.nameFileLine()

	if file != "" {
		file = cropFilename(file, name)
	}

	putCache(l, nfl{
		name: name,
		file: file,
		line: line,
	})

	return
}

func (l PC) nameFileLine() (name, file string, line int) {
	// Use pool to avoid allocation
	p := framesPool.Get().(*[]uintptr)
	defer framesPool.Put(p)

	(*p)[0] = uintptr(l)
	fs := runtime.CallersFrames(*p)
	f, _ := fs.Next()
	return f.Function, f.File, f.Line
}

func (l PC) FuncEntry() PC {
	if l == 0 {
		return 0
	}

	f := runtime.FuncForPC(uintptr(l))
	if f == nil {
		return 0
	}

	return PC(f.Entry())
}

// Caller returns information about the calling goroutine's stack. The argument s is the number of frames to ascend, with 0 identifying the caller of Caller.
//
// It's hacked version of runtime.Caller with no allocs.
func Caller(s int) (r PC) {
	caller1(1+s, &r, 1, 1)

	return
}

// FuncEntry returns information about the calling goroutine's stack. The argument s is the number of frames to ascend, with 0 identifying the caller of Caller.
//
// It's hacked version of runtime.Callers -> runtime.CallersFrames -> Frames.Next -> Frame.Entry with no allocs.
func FuncEntry(s int) (r PC) {
	caller1(1+s, &r, 1, 1)

	return r.FuncEntry()
}

func CallerOnce(s int, pc *PC) (r PC) {
	r = PC(atomic.LoadUintptr((*uintptr)(unsafe.Pointer(pc))))
	if r != 0 {
		return
	}

	caller1(1+s, &r, 1, 1)

	atomic.StoreUintptr((*uintptr)(unsafe.Pointer(pc)), uintptr(r))

	return
}

func FuncEntryOnce(s int, pc *PC) (r PC) {
	r = PC(atomic.LoadUintptr((*uintptr)(unsafe.Pointer(pc))))
	if r != 0 {
		return
	}

	caller1(1+s, &r, 1, 1)

	r = r.FuncEntry()

	atomic.StoreUintptr((*uintptr)(unsafe.Pointer(pc)), uintptr(r))

	return
}

// Callers returns callers stack trace.
//
// It's hacked version of runtime.Callers -> runtime.CallersFrames -> Frames.Next -> Frame.Entry with only one alloc (resulting slice).
func Callers(skip, n int) PCs {
	// Try to reuse from pool
	if n <= 64 {
		p := pcsPool.Get().(*PCs)
		tr := (*p)[:n]
		n = callers(1+skip, tr)

		// Copy result to new slice of exact size
		res := make(PCs, n)
		copy(res, tr[:n])

		pcsPool.Put(p)
		return res
	}

	// Fallback for large requests
	tr := make(PCs, n)
	n = callers(1+skip, tr)
	return tr[:n]
}

// CallersFill puts callers stack trace into provided slice.
//
// It's hacked version of runtime.Callers -> runtime.CallersFrames -> Frames.Next -> Frame.Entry with no allocs.
func CallersFill(skip int, tr PCs) PCs {
	n := callers(1+skip, tr)
	return tr[:n]
}

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
