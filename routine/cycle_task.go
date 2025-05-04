package routine

import (
	"context"
	"sync/atomic"
	"time"
)

// CycleTask represents a recurring task
type CycleTask struct {
	t        Task
	interval time.Duration
	ticker   *time.Ticker
	running  atomic.Bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewCycleTask creates an instance of RecurringTask
func NewCycleTask(t Task, i time.Duration) *CycleTask {
	rt := &CycleTask{
		t:        t,
		interval: i,
	}
	return rt
}

// Start starts the timer
func (t *CycleTask) Start(ctx context.Context) error {
	if t.running.Load() {
		return ErrTaskAlreadyRunning
	}
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.ticker = time.NewTicker(t.interval)
	ready := make(chan struct{})
	go func() {
		close(ready)
		for {
			select {
			case <-t.ctx.Done():
				return
			case <-t.ticker.C:
				t.t(t.ctx)
			}
		}
	}()

	<-ready
	t.running.Store(true)
	return nil
}

// Stop stops the timer
func (t *CycleTask) Stop(_ context.Context) error {
	if !t.running.Load() {
		return nil
	}
	t.cancel()
	if t.ticker != nil {
		t.ticker.Stop()
	}
	t.running.Store(false)
	return nil
}
