package accumulator

import (
	"time"
)

type sliding struct {
	previous     *window
	current      *window
	limit        uint64
	windowLength time.Duration
}

func newSliding(limit uint64, windowLength time.Duration) *sliding {
	zero := time.Unix(0, 0)
	return &sliding{
		previous:     newWindow(0, zero),
		current:      newWindow(0, zero),
		limit:        limit,
		windowLength: windowLength,
	}
}
