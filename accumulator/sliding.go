package accumulator

type sliding struct {
	previous *window
	current  *window
	limit    uint64
	seconds  uint64
}

func newSliding(limit uint64, seconds uint64) *sliding {
	return &sliding{
		previous: newWindow(0, 0),
		current:  newWindow(0, 0),
		limit:    limit,
		seconds:  seconds,
	}
}
