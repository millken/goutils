package routine

import (
	"context"
	"testing"
	"time"
)

func TestDelayTaskTimeout(t *testing.T) {
	c := make(chan bool)
	ctx := context.Background()
	task := NewDelayTask(func(ctx context.Context) error { c <- true; return nil }, 100*time.Millisecond)
	task.Start(ctx)
	if err := task.Start(ctx); err != ErrTaskAlreadyRunning {
		t.Fail()
	}
	defer func() {
		task.Stop(ctx)
	}()

	ret := <-c
	if !ret {
		t.Fail()
	}
}

func TestDelayTaskStop(t *testing.T) {
	c := make(chan bool)
	ctx := context.Background()
	task := NewDelayTask(func(ctx context.Context) error { c <- true; return nil }, 100*time.Millisecond)
	task.Start(ctx)
	task.Stop(ctx)

	select {
	case <-c:
		t.Fail()
	case <-time.After(600 * time.Millisecond):
	}
}
