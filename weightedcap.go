package weightedcap

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type weightedCap struct {
	capacity    int64
	maxCapacity int64

	signalMu sync.Mutex
	signal   chan struct{}
}

func New(capacity int64) *weightedCap {
	return &weightedCap{
		maxCapacity: capacity,
		capacity:    capacity,
		signal:      make(chan struct{}, 1),
	}
}

func (w *weightedCap) Consume(ctx context.Context, n int64) error {
	if n > w.maxCapacity {
		return &ExceedingCapacityErr{n, w.maxCapacity}
	}
	if atomic.LoadInt64(&w.capacity) >= n {
		w.consume(ctx, n)
		return nil
	}

	w.signalMu.Lock()
	defer w.signalMu.Unlock()

	for atomic.LoadInt64(&w.capacity) < n {
		err := w.waitForSignalLocked(ctx)
		if err != nil {
			return err
		}
	}

	w.consume(ctx, n)
	return nil
}

func (w *weightedCap) waitForSignalLocked(ctx context.Context) error {
	select {
	case <-w.signal:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *weightedCap) consume(ctx context.Context, n int64) {
	atomic.AddInt64(&w.capacity, -n)
}

func (w *weightedCap) Release(n int64) {
	atomic.AddInt64(&w.capacity, n)
	w.signal <- struct{}{}
}

type ExceedingCapacityErr struct {
	Capacity    int64
	MaxCapacity int64
}

func (e *ExceedingCapacityErr) Error() string {
	return fmt.Sprintf("weightedcap: requested capacity (%d) higher than the maximum capacity (%d)", e.Capacity, e.MaxCapacity)
}
