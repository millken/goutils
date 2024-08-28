package accumulator

import (
	"goutils/fasttime"
	"hash/maphash"
	"strconv"
	"time"
)

var instance *Accumulator = NewAccumulator()
var mapseed = maphash.MakeSeed()

type Hasher func(string) uint64

func defaultHasher(v string) uint64 {
	return maphash.String(mapseed, v)
}

type Accumulator struct {
	buckets [bucketsCount]bucket
	hasher  Hasher
}

func NewAccumulator() *Accumulator {
	acc := &Accumulator{
		hasher: defaultHasher,
	}
	for i := range acc.buckets[:] {
		acc.buckets[i].Reset()
	}
	return acc
}

func (a *Accumulator) AllowN(key string, n uint64, limit uint64, windowLength time.Duration) bool {
	hash := a.hashKey(key, limit, windowLength)
	idx := hash % bucketsCount

	entry, exists := a.buckets[idx].Get(hash)
	if !exists {
		entry = newEntry(limit, windowLength)
		a.buckets[idx].Set(hash, entry)
	}

	currentTime := fasttime.Now().UTC()
	sizeAlignedTime := currentTime.Truncate(windowLength)
	timeSinceStart := sizeAlignedTime.Sub(entry.current.getStartTime())
	nSlides := uint64(timeSinceStart / windowLength)
	sizeAlignedTime2 := sizeAlignedTime.Add(-windowLength)

	// 处理窗口滑动
	if nSlides == 1 {
		// 如果窗口滑动一次，将当前窗口的计数复制到前一个窗口，并重置当前窗口
		entry.previous.setToState(sizeAlignedTime2, entry.current.count)
		entry.current.resetToTime(sizeAlignedTime)
	} else if nSlides > 1 {
		// 如果窗口滑动超过一次，重置前一个窗口和当前窗口
		entry.previous.resetToTime(sizeAlignedTime2)
		entry.current.resetToTime(sizeAlignedTime)
	}

	currentWindowStart := entry.current.getStartTime()
	currentWindowBoundary := currentTime.Sub(currentWindowStart)

	w := float64(windowLength-currentWindowBoundary) / float64(windowLength)
	currentSlidingRequests := uint64(w*float64(entry.previous.count)) + entry.current.count

	if currentSlidingRequests+n > entry.limit {
		return false
	}

	entry.current.updateCount(n)
	return true
}

func (a *Accumulator) hashKey(key string, limit uint64, windowLength time.Duration) uint64 {
	//TODO: use a better hashing algorithm
	key = key + strconv.FormatUint(limit, 10) + windowLength.String()
	return a.hasher(key)
}

// Allow returns true if the request is allowed, otherwise false.
// The key is the unique identifier of the request, limit is the maximum
// number of requests allowed in the duration, and size is the duration
func Allow(key string, limit uint64, windowLength time.Duration) bool {
	return instance.AllowN(key, 1, limit, windowLength)
}

func AllowN(key string, n uint64, limit uint64, windowLength time.Duration) bool {
	return instance.AllowN(key, n, limit, windowLength)
}
