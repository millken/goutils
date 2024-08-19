package accumulator

import (
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
	e, ok := a.buckets[idx].Get(hash)
	if !ok {
		e = newEntry(limit, windowLength)
		a.buckets[idx].Set(hash, e)
	}
	currentTime := time.Now().UTC()

	sizeAlignedTime := currentTime.Truncate(windowLength)
	timeSinceStart := sizeAlignedTime.Sub(e.current.getStartTime())
	nSlides := timeSinceStart / windowLength
	sizeAlignedTime2 := sizeAlignedTime.Add(-windowLength)

	// window slide shares both current and previous windows.
	if nSlides == 1 {
		e.previous.setToState(sizeAlignedTime2, e.current.count)
		e.current.resetToTime(sizeAlignedTime)

	} else if nSlides > 1 {
		e.previous.resetToTime(sizeAlignedTime2)
		e.current.resetToTime(sizeAlignedTime)
	}

	currentWindowBoundary := currentTime.Sub(e.current.getStartTime())
	w := float64(e.windowLength-currentWindowBoundary) / float64(e.windowLength)
	currentSlidingRequests := uint64(w*float64(e.previous.count)) + e.current.count
	if currentSlidingRequests+n > e.limit {
		return false
	}
	// diff := currentTime.Sub(sizeAlignedTime)
	// rate := float64(entry.previous.count)*(float64(entry.windowLength)-float64(diff))/float64(entry.windowLength) + float64(entry.current.count)
	// nrate := uint64(math.Round(rate))
	// if nrate >= entry.limit {
	// 	return false
	// }

	// add current request count to window of current count
	e.current.updateCount(n)
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
