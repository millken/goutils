package faststack

import (
	"path"
	"path/filepath"
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
	assert.Equal(t, 18, line)
}

func TestLocationShort(t *testing.T) {
	pc := Caller(0)
	assert.Equal(t, "faststack_test.go:26", pc.String())
}

func TestLocation2(t *testing.T) {
	func() {
		func() {
			l := FuncEntry(0)

			assert.Equal(t, "faststack_test.go:32", l.String())
		}()
	}()
}

func TestLocationOnce(t *testing.T) {
	var pc PC

	CallerOnce(-1, &pc)
	assert.Equal(t, "faststack.go:112", pc.String())

	pc++
	save := pc

	CallerOnce(-1, &pc)

	assert.Equal(t, save, pc) // not changed

	//
	pc = 0

	FuncEntryOnce(-1, &pc)
	assert.Equal(t, "faststack.go:119", pc.String())

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
	assert.Equal(t, "faststack_test.go:97", st[0].String())
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
	assert.Equal(t, "faststack_test.go:107", st[0].String())
	assert.Equal(t, "faststack_test.go:108", st[1].String())
	assert.Equal(t, "faststack_test.go:117", st[2].String())

	re := `faststack_test.go:107 at faststack_test.go:108 at faststack_test.go:117`

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
	assert.Equal(t, "faststack_test.go:138", l.String())
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
