package accumulator

import (
	"goutils/fasttime"
	"hash/maphash"
	"strconv"
	"sync"
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

func (a *Accumulator) AllowN(key string, n uint64, limit uint64, seconds uint64) bool {
	hash := a.hashKey(key, limit, seconds)
	sliding, ok := a.getsliding(hash)
	if !ok {
		sliding = newSliding(limit, seconds)
		a.mu.Lock()
		a.slidings[hash] = sliding
		a.mu.Unlock()
	}
	currentTime := fasttime.UnixTimestamp()

	sizeAlignedTime := currentTime - (currentTime % seconds)
	timeSinceStart := sizeAlignedTime - sliding.current.getStartTime()
	nSlides := timeSinceStart / seconds

	// window slide shares both current and previous windows.
	if nSlides == 1 {
		sliding.previous.setToState(sizeAlignedTime-seconds, sliding.current.count)
		sliding.current.resetToTime(sizeAlignedTime)

	} else if nSlides > 1 {
		sliding.previous.resetToTime(sizeAlignedTime - seconds)
		sliding.current.resetToTime(sizeAlignedTime)
	}

	currentWindowBoundary := currentTime - sliding.current.getStartTime()

	w := float64(sliding.seconds-currentWindowBoundary) / float64(sliding.seconds)

	currentSlidingRequests := uint64(w*float64(sliding.previous.count)) + sliding.current.count

	// diff := currentTime - sizeAlignedTime
	// rate := float64(sliding.previous.count)*(float64(sliding.seconds)-float64(diff))/float64(sliding.seconds) + float64(sliding.current.count)
	// fmt.Println("rate", rate)
	if currentSlidingRequests+n > sliding.limit {
		return false
	}

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

func (a *Accumulator) hashKey(key string, limit uint64, seconds uint64) uint64 {
	key = key + strconv.FormatUint(limit, 10) + strconv.FormatUint(seconds, 10)
	return a.hasher(key)
}

// Allow returns true if the request is allowed, otherwise false.
// The key is the unique identifier of the request, limit is the maximum
// number of requests allowed in the duration, and size is the duration
func Allow(key string, limit uint64, seconds uint64) bool {
	return instance.AllowN(key, 1, limit, seconds)
}

func AllowN(key string, n uint64, limit uint64, seconds uint64) bool {
	return instance.AllowN(key, n, limit, seconds)
}
