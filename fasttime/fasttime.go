package fasttime

import (
	"sync/atomic"
	"time"
)

var (
	correctionDur time.Duration = time.Millisecond * 100
	dur           time.Duration = time.Millisecond * 5
	_t            atomic.Pointer[time.Time]
)

func init() {
	ticker := time.Tick(dur)
	lastCorrection := time.Now()
	_t.Store(&lastCorrection)
	go func() {
		for {
			t := <-ticker
			// rely on ticker for approximation
			if t.Sub(lastCorrection) < correctionDur {
				now := Now().Add(dur)
				_t.Store(&now)
			} else { // correct the  time at a fixed interval
				now := time.Now()
				_t.Store(&now)
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
	return *_t.Load()
}

func Now() time.Time {
	return Time()
}
