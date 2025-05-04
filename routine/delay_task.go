package routine

import (
	"context"
	"sync/atomic"
	"time"
)

// DelayTask represents a timeout task
type DelayTask struct {
	cb       Task
	duration time.Duration
	running  atomic.Bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewDelayTask creates an instance of DelayTask
func NewDelayTask(cb Task, d time.Duration) *DelayTask {
	dt := &DelayTask{
		cb:       cb,
		duration: d,
	}
	return dt
}

// Start executes the delayed task after given timeout.
func (t *DelayTask) Start(ctx context.Context) error {
	if t.running.Load() {
		return ErrTaskAlreadyRunning
	}
	t.ctx, t.cancel = context.WithCancel(ctx)
	ready := make(chan struct{})
	go func() {
		close(ready)
		select {
		case <-t.ctx.Done():
			return
		case <-time.After(t.duration):
			t.cb(t.ctx)
		}
	}()

	<-ready
	t.running.Store(true)
	return nil
}

// Stop stops the timeout
func (t *DelayTask) Stop(ctx context.Context) error {
	if !t.running.Load() {
		return nil
	}
	t.cancel()
	t.running.Store(false)
	return nil
}
