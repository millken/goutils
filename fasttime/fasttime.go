package fasttime

import (
	"sync/atomic"
	"time"
)

var (
	lastCorrection               = time.Now()
	correctionDur  time.Duration = time.Millisecond * 100
	dur            time.Duration = time.Millisecond * 5
	_dur           int64
	_t             atomic.Pointer[time.Time]
)

func init() {
	atomic.StoreInt64(&_dur, dur.Nanoseconds())
	ticker := time.Tick(time.Duration(atomic.LoadInt64(&_dur)))
	_t.Store(&lastCorrection)
	go func() {
		for atomic.LoadInt64(&_dur) > 0 {
			t := <-ticker
			// rely on ticker for approximation
			if t.Sub(lastCorrection) < correctionDur {
				now := Now().Add(time.Duration(atomic.LoadInt64(&_dur)))
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
