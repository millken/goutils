package sqldb

import (
	"strconv"
	"strings"
	"time"
	_ "unsafe" // required to use //go:linkname
)

//go:noescape
//go:linkname nanotime runtime.nanotime
func nanotime() int64
func Now() uint64 {
	return uint64(nanotime())
}

// Since returns the amount of time that has elapsed since t. t should be
// the result of a call to Now() on the same machine.
func Since(t uint64) time.Duration {
	return time.Duration(Now() - t)
}

// Supported drivers.
const (
	invalidFlavor Flavor = iota

	MySQL
	PostgreSQL
	SQLite
)

// Flavor is the flag to control the format of compiled sql.
type Flavor int

// String returns the name of f.
func (f Flavor) String() string {
	switch f {
	case MySQL:
		return "MySQL"
	case PostgreSQL:
		return "PostgreSQL"
	case SQLite:
		return "SQLite"
	}

	return "<invalid>"
}

func (f Flavor) tableQuote(prefix string, table string) string {
	tableQuote := "`"
	switch f {
	case PostgreSQL:
		tableQuote = "\""
	}

	if strings.Contains(table, ".") {
		return tableQuote + strings.ReplaceAll(table, ".", tableQuote+"."+tableQuote) + tableQuote
	}

	return tableQuote + prefix + table + tableQuote
}

func (f Flavor) columnQuote(column string) string {
	columnQuote := ""
	switch f {
	case PostgreSQL:
		columnQuote = "\""
	default:
		columnQuote = "`"
	}
	if strings.ContainsRune(column, '.') {
		if strings.ContainsRune(column, '*') {
			return columnQuote + strings.ReplaceAll(column, ".", columnQuote+".")
		}
		return columnQuote + strings.ReplaceAll(column, ".", columnQuote+"."+columnQuote) + columnQuote
	} else if strings.Contains(column, "(") || strings.Contains(column, " ") {
		return column
	}

	return columnQuote + column + columnQuote
}

func (f Flavor) placeHolder(dataLen int) string {
	switch f {
	case MySQL, SQLite:
		return strings.Repeat("?,", dataLen)[:dataLen*2-1]
	case PostgreSQL:
		var placeholder string
		for i := 1; i <= dataLen; i++ {
			placeholder += "$" + strconv.Itoa(i) + ","
		}
		return placeholder[:len(placeholder)-1]
	}
	return ""
}
