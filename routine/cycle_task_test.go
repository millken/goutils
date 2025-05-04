package routine

import (
	"context"
	"sync"
	"testing"
	"time"
)

type MockHandler struct {
	Count uint
	mu    sync.RWMutex
}

func (h *MockHandler) Do(ctx context.Context) error {
	h.mu.Lock()
	h.Count++
	h.mu.Unlock()
	return nil
}

func TestRecurringTask(t *testing.T) {
	h := &MockHandler{Count: 0}
	ctx := context.Background()
	task := NewCycleTask(h.Do, 100*time.Millisecond)
	task.Start(ctx)
	if err := task.Start(ctx); err != ErrTaskAlreadyRunning {
		t.Fail()
	}
	time.Sleep(600 * time.Millisecond)
	task.Stop(ctx)
	h.mu.RLock()
	if h.Count < 5 {
		t.Fail()
	}
	h.mu.RUnlock()
}
