package accumulator

// window represents the structure of timing-window at given point of time.
type window struct {
	count     uint64
	startTime uint64
}

func (w *window) updateCount(n uint64) {
	w.count += n
}

func (w *window) getStartTime() uint64 {
	return w.startTime
}

func (w *window) setStateFrom(other *window) {
	w.count = other.count
	w.startTime = other.startTime
}

func (w *window) resetToTime(startTime uint64) {
	w.count = 0
	w.startTime = startTime
}

func (w *window) setToState(startTime uint64, count uint64) {
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
func newWindow(count uint64, startTime uint64) *window {

	return &window{
		count:     count,
		startTime: startTime,
	}
}