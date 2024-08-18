package accumulator

import "time"

// window represents the structure of timing-window at given point of time.
type window struct {
	count     uint64
	startTime time.Time
}

func (w *window) updateCount(n uint64) {
	w.count += n
}

func (w *window) getStartTime() time.Time {
	return w.startTime
}

func (w *window) setStateFrom(other *window) {
	w.count = other.count
	w.startTime = other.startTime
}

func (w *window) resetToTime(startTime time.Time) {
	w.count = 0
	w.startTime = startTime
}

func (w *window) setToState(startTime time.Time, count uint64) {
	w.startTime = startTime
	w.count = count
}

// Creates and returns a pointer to the new Window instance.
//
// Parameters:
//
// 1. count: The initial count of the window.
//
// 2. startTime: The initial starting time of the window.
func newWindow(count uint64, startTime time.Time) *window {

	return &window{
		count:     count,
		startTime: startTime,
	}
}
