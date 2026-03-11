package faststack

import (
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestLocation(t *testing.T) {
	testLocationInside(t)
}

func testLocationInside(t *testing.T) {
	t.Helper()

	pc := Caller(0)
	name, file, line := pc.NameFileLine()
	assert.Equal(t, "faststack.testLocationInside", path.Base(name))
	assert.Equal(t, "faststack_test.go", filepath.Base(file))
	assert.Equal(t, 19, line)
}

func TestLocationShort(t *testing.T) {
	pc := Caller(0)
	assert.Equal(t, "faststack_test.go:27", pc.String())
}

func TestLocation2(t *testing.T) {
	func() {
		func() {
			l := FuncEntry(0)

			assert.Equal(t, "faststack_test.go:31", l.String())
		}()
	}()
}

func TestLocationOnce(t *testing.T) {
	var pc PC

	CallerOnce(-1, &pc)
	assert.Equal(t, "faststack.go:185", pc.String())

	pc++
	save := pc

	CallerOnce(-1, &pc)

	assert.Equal(t, save, pc) // not changed

	//
	pc = 0

	FuncEntryOnce(-1, &pc)
	assert.Equal(t, "faststack.go:192", pc.String())

	pc++
	save = pc

	FuncEntryOnce(-1, &pc)

	assert.Equal(t, save, pc) // not changed
}

func TestLocationCropFileName(t *testing.T) {
	assert.Equal(t, "github.com/nikandfor/tlog/sub/module/file.go",
		cropFilename("/path/to/src/github.com/nikandfor/tlog/sub/module/file.go", "github.com/nikandfor/tlog/sub/module.(*type).method"))
	assert.Equal(t, "github.com/nikandfor/tlog/sub/module/file.go",
		cropFilename("/path/to/src/github.com/nikandfor/tlog/sub/module/file.go", "github.com/nikandfor/tlog/sub/module.method"))
	assert.Equal(t, "github.com/nikandfor/tlog/root.go", cropFilename("/path/to/src/github.com/nikandfor/tlog/root.go", "github.com/nikandfor/tlog.type.method"))
	assert.Equal(t, "github.com/nikandfor/tlog/root.go", cropFilename("/path/to/src/github.com/nikandfor/tlog/root.go", "github.com/nikandfor/tlog.method"))
	assert.Equal(t, "root.go", cropFilename("/path/to/src/root.go", "github.com/nikandfor/tlog.method"))
	assert.Equal(t, "sub/file.go", cropFilename("/path/to/src/sub/file.go", "github.com/nikandfor/tlog/sub.method"))
	assert.Equal(t, "root.go", cropFilename("/path/to/src/root.go", "tlog.method"))
	assert.Equal(t, "subpkg/file.go", cropFilename("/path/to/src/subpkg/file.go", "subpkg.method"))
	assert.Equal(t, "subpkg/file.go", cropFilename("/path/to/src/subpkg/file.go", "github.com/nikandfor/tlog/subpkg.(*type).method"))
	assert.Equal(t, "errors/fmt_test.go",
		cropFilename("/home/runner/work/errors/errors/fmt_test.go", "tlog.app/go/error.TestErrorFormatCaller"))
	assert.Equal(t, "jq/object_test.go", cropFilename("/Users/nik/nikandfor/jq/object_test.go", "nikand.dev/go/jq.TestObject"))

	// Edge cases: no dot after slash (e.g., package.Function without type receiver)
	// Tests the fix for pp == -1 handling
	assert.Equal(t, "myapp/file.go", cropFilename("/path/to/myapp/file.go", "myapp.Function"))
	assert.Equal(t, "myapp/sub/file.go", cropFilename("/path/to/myapp/sub/file.go", "myapp/sub.Function"))
}

func TestCaller(t *testing.T) {
	a, b := Caller(0),
		Caller(0)

	//	assert.False(t, a == b, "%x == %x", uintptr(a), uintptr(b))
	assert.NotEqual(t, a, b)
}

// align line numbers for tests

func TestLocationFillCallers(t *testing.T) {
	st := make(PCs, 1)

	st = CallersFill(0, st)

	assert.Equal(t, 1, len(st))
	assert.Equal(t, "faststack_test.go:103", st[0].String())
}

func testLocationsInside() (st PCs) {
	func() {
		func() {
			st = Callers(1, 3)
		}()
	}()

	return
}

func TestLocationPCsString(t *testing.T) {
	var st PCs
	func() {
		func() {
			st = testLocationsInside()
		}()
	}()

	assert.Equal(t, 3, len(st))
	assert.Equal(t, "faststack_test.go:113", st[0].String())
	assert.Equal(t, "faststack_test.go:114", st[1].String())
	assert.Equal(t, "faststack_test.go:123", st[2].String())

	re := `faststack_test.go:113 at faststack_test.go:114 at faststack_test.go:123`

	assert.Equal(t, re, st.String())
}

func TestLocation3(t *testing.T) {
	testInline(t)
}

func testInline(t *testing.T) {
	t.Helper()

	testLocation3(t)
}

func testLocation3(t *testing.T) {
	t.Helper()

	l := Caller(1)
	assert.Equal(t, "faststack_test.go:144", l.String())
}

func TestLocationZero(t *testing.T) {
	var l PC

	entry := l.FuncEntry()
	assert.Equal(t, PC(0), entry)

	entry = PC(100).FuncEntry()
	assert.Equal(t, PC(0), entry)

	name, file, line := l.NameFileLine()
	assert.Equal(t, "", name)
	assert.Equal(t, "", file)
	assert.Equal(t, 0, line)
}

func BenchmarkLocationCaller(b *testing.B) {
	b.ReportAllocs()

	var l PC

	for i := 0; i < b.N; i++ {
		l = Caller(0)
	}

	_ = l
}

func BenchmarkLocationNameFileLine(b *testing.B) {
	b.ReportAllocs()

	var n, f string
	var line int

	l := Caller(0)

	for i := 0; i < b.N; i++ {
		n, f, line = l.nameFileLine()
	}

	_, _, _ = n, f, line //nolint:dogsled
}

func TestCache(t *testing.T) {
	// Test the new sync.Map based cache
	pc1 := PC(1000)
	pc2 := PC(2000)

	putCache(pc1, nfl{name: "func1", file: "file1.go", line: 10})
	putCache(pc2, nfl{name: "func2", file: "file2.go", line: 20})

	// Both should be present
	v, ok := getCache(pc1)
	assert.True(t, ok)
	assert.Equal(t, "func1", v.name)

	v, ok = getCache(pc2)
	assert.True(t, ok)
	assert.Equal(t, "func2", v.name)

	// Update existing entry
	putCache(pc1, nfl{name: "func1_updated", file: "file1_updated.go", line: 99})
	v, ok = getCache(pc1)
	assert.True(t, ok)
	assert.Equal(t, "func1_updated", v.name)
}

func TestCacheMiss(t *testing.T) {
	// Non-existent PC should return miss
	_, ok := getCache(PC(99999))
	assert.False(t, ok)
}

// ============================================================================
// Performance Comparison Benchmarks: faststack vs runtime
// ============================================================================

// BenchmarkRuntimeCaller tests runtime.Caller performance
func BenchmarkRuntimeCaller(b *testing.B) {
	b.ReportAllocs()

	var pc uintptr

	for i := 0; i < b.N; i++ {
		pc, _, _, _ = runtime.Caller(0)
	}

	_ = pc
}

// BenchmarkFaststackCaller tests faststack.Caller performance
func BenchmarkFaststackCaller(b *testing.B) {
	b.ReportAllocs()

	var pc PC

	for i := 0; i < b.N; i++ {
		pc = Caller(0)
	}

	_ = pc
}

// BenchmarkRuntimeCallerFrames tests runtime.CallersFrames performance
func BenchmarkRuntimeCallerFrames(b *testing.B) {
	b.ReportAllocs()

	var name, file string
	var line int

	// Pre-capture PC
	pcs := make([]uintptr, 1)
	runtime.Callers(0, pcs)
	pc := pcs[0]

	for i := 0; i < b.N; i++ {
		frames := runtime.CallersFrames([]uintptr{pc})
		f, _ := frames.Next()
		name, file, line = f.Function, f.File, f.Line
	}

	_, _, _ = name, file, line
}

// BenchmarkFaststackNameFileLine tests faststack.NameFileLine with cache
func BenchmarkFaststackNameFileLine(b *testing.B) {
	b.ReportAllocs()

	var name, file string
	var line int

	pc := Caller(0)

	for i := 0; i < b.N; i++ {
		name, file, line = pc.NameFileLine()
	}

	_, _, _ = name, file, line
}

// BenchmarkFaststackNameFileLineNoCache tests without cache (raw runtime call)
func BenchmarkFaststackNameFileLineNoCache(b *testing.B) {
	b.ReportAllocs()

	var name, file string
	var line int

	pc := Caller(0)

	for i := 0; i < b.N; i++ {
		name, file, line = pc.nameFileLine()
	}

	_, _, _ = name, file, line
}

// BenchmarkRuntimeFuncForPC tests runtime.FuncForPC performance
func BenchmarkRuntimeFuncForPC(b *testing.B) {
	b.ReportAllocs()

	var pc uintptr

	pcs := make([]uintptr, 1)
	runtime.Callers(0, pcs)
	rpc := pcs[0]

	for i := 0; i < b.N; i++ {
		f := runtime.FuncForPC(rpc)
		if f != nil {
			pc = f.Entry()
		}
	}

	_ = pc
}

// BenchmarkFaststackFuncEntry tests faststack.FuncEntry performance
func BenchmarkFaststackFuncEntry(b *testing.B) {
	b.ReportAllocs()

	var pc PC

	rpc := Caller(0)

	for i := 0; i < b.N; i++ {
		pc = rpc.FuncEntry()
	}

	_ = pc
}

// BenchmarkFaststackCallers tests faststack.Callers (capturing 4 frames)
func BenchmarkFaststackCallers4(b *testing.B) {
	b.ReportAllocs()

	var pcs PCs

	for i := 0; i < b.N; i++ {
		pcs = Callers(0, 4)
	}

	_ = pcs
}

// BenchmarkRuntimeCallers tests runtime.Callers (capturing 4 frames)
func BenchmarkRuntimeCallers4(b *testing.B) {
	b.ReportAllocs()

	var pcs []uintptr

	for i := 0; i < b.N; i++ {
		pcs = make([]uintptr, 4)
		runtime.Callers(0, pcs)
	}

	_ = pcs
}

// BenchmarkFaststackString tests PC.String() formatting
func BenchmarkFaststackString(b *testing.B) {
	b.ReportAllocs()

	var s string

	pc := Caller(0)

	for i := 0; i < b.N; i++ {
		s = pc.String()
	}

	_ = s
}

// BenchmarkPCsString tests PCs.String() formatting (3 frames)
func BenchmarkPCsString(b *testing.B) {
	b.ReportAllocs()

	var s string

	pcs := Callers(0, 3)

	for i := 0; i < b.N; i++ {
		s = pcs.String()
	}

	_ = s
}
