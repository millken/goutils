package fasttime

import (
	"sync/atomic"
	"time"
	_ "unsafe"
)

var (
	correctionDur time.Duration = time.Millisecond * 100
	dur           time.Duration = time.Millisecond * 20
	pt            atomic.Pointer[time.Time]
)

//go:linkname Now1 time.now
func Now1() (sec int64, nsec int32, mono int64)

// DateClock is faster version of t.Date(); t.Clock().
func DateClock(t time.Time) (year, month, day, hour, min, sec int) { //nolint:gocritic
	u := timeAbs(t)
	year, month, day, _ = absDate(u, true)
	hour, min, sec = absClock(u)
	return
}

//go:linkname timeAbs time.Time.abs
func timeAbs(time.Time) uint64

//go:linkname absClock time.absClock
func absClock(uint64) (hour, min, sec int)

//go:linkname absDate time.absDate
func absDate(uint64, bool) (year, month, day, yday int)

func init() {
	ticker := time.Tick(dur)
	lastCorrection := time.Now()
	pt.Store(&lastCorrection)
	go func() {
		for {
			t := <-ticker
			// rely on ticker for approximation
			if t.Sub(lastCorrection) < correctionDur {
				now := Now().Add(dur)
				pt.Store(&now)
			} else { // correct the  time at a fixed interval
				now := time.Now()
				pt.Store(&now)
				lastCorrection = t
			}
		}
	}()
}

// UnixTimestamp returns the current unix timestamp in seconds.
//
// It is faster than time.Now().Unix()
func UnixTimestamp() int64 {
	return Now().Unix()
}

// UnixDate returns date from the current unix timestamp.
//
// The date is calculated by dividing unix timestamp by (24*3600)
func UnixDate() int64 {
	return UnixTimestamp() / (24 * 3600)
}

// UnixHour returns hour from the current unix timestamp.
//
// The hour is calculated by dividing unix timestamp by 3600
func UnixHour() int64 {
	return UnixTimestamp() / 3600
}

// UnixMinute returns minute from the current unix timestamp.
func UnixMinute() int64 {
	return UnixTimestamp() / 60
}

// Time returns the current time.Time
func Time() time.Time {
	return *pt.Load()
}

func Now() time.Time {
	return Time()
}
