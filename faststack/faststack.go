package faststack

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

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
	locmu sync.RWMutex
	locc  = map[PC]nfl{}
)

// NameFileLine returns function name, file and line number for location.
//
// This works only in the same binary where location was captured.
//
// This functions is a little bit modified version of runtime.(*Frames).Next().
func (l PC) NameFileLine() (name, file string, line int) {
	if l == 0 {
		return
	}

	locmu.RLock()
	c, ok := locc[l]
	locmu.RUnlock()
	if ok {
		return c.name, c.file, c.line
	}

	name, file, line = l.nameFileLine()

	if file != "" {
		file = cropFilename(file, name)
	}

	locmu.Lock()
	locc[l] = nfl{
		name: name,
		file: file,
		line: line,
	}
	locmu.Unlock()

	return
}

func (l PC) nameFileLine() (name, file string, line int) {
	fs := runtime.CallersFrames([]uintptr{uintptr(l)})
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
	tr := make([]PC, n)
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
	tp = tp[:p+1+pp] // cut type and func name

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
