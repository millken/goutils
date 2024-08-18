package accumulator

import (
	"hash/maphash"
	"strconv"
	"sync"
	"time"
)

var instance *Accumulator = NewAccumulator()
var mapseed = maphash.MakeSeed()

type Hasher func(string) uint64

func defaultHasher(v string) uint64 {
	return maphash.String(mapseed, v)
}

type Accumulator struct {
	mu       sync.RWMutex
	slidings map[uint64]*sliding
	hasher   Hasher
}

func NewAccumulator() *Accumulator {
	return &Accumulator{
		slidings: make(map[uint64]*sliding),
		hasher:   defaultHasher,
	}
}

func (a *Accumulator) AllowN(key string, n uint64, limit uint64, windowLength time.Duration) bool {
	hash := a.hashKey(key, limit, windowLength)
	sliding, ok := a.getsliding(hash)
	if !ok {
		sliding = newSliding(limit, windowLength)
		a.mu.Lock()
		a.slidings[hash] = sliding
		a.mu.Unlock()
	}
	currentTime := time.Now().UTC()

	sizeAlignedTime := currentTime.Truncate(windowLength)
	timeSinceStart := sizeAlignedTime.Sub(sliding.current.getStartTime())
	nSlides := timeSinceStart / windowLength
	sizeAlignedTime2 := sizeAlignedTime.Add(-windowLength)

	// window slide shares both current and previous windows.
	if nSlides == 1 {
		sliding.previous.setToState(sizeAlignedTime2, sliding.current.count)
		sliding.current.resetToTime(sizeAlignedTime)

	} else if nSlides > 1 {
		sliding.previous.resetToTime(sizeAlignedTime2)
		sliding.current.resetToTime(sizeAlignedTime)
	}

	currentWindowBoundary := currentTime.Sub(sliding.current.getStartTime())
	w := float64(sliding.windowLength-currentWindowBoundary) / float64(sliding.windowLength)
	currentSlidingRequests := uint64(w*float64(sliding.previous.count)) + sliding.current.count
	if currentSlidingRequests+n > sliding.limit {
		return false
	}
	// diff := currentTime.Sub(sizeAlignedTime)
	// rate := float64(sliding.previous.count)*(float64(sliding.windowLength)-float64(diff))/float64(sliding.windowLength) + float64(sliding.current.count)
	// nrate := uint64(math.Round(rate))
	// if nrate >= sliding.limit {
	// 	return false
	// }

	// add current request count to window of current count
	sliding.current.updateCount(n)
	return true
}

func (a *Accumulator) getsliding(hash uint64) (*sliding, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	sliding, ok := a.slidings[hash]
	return sliding, ok
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
