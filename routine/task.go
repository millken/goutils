package routine

import (
	"context"
	"errors"
)

var (
	ErrTaskAlreadyRunning = errors.New("task already running")
)

type Task func(ctx context.Context) error
