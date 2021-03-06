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

type Cap interface {
	Consume(ctx context.Context, n int64) (release func(), err error)
}

func New(capacity int64) *weightedCap {
	cap := &weightedCap{
		maxCapacity: capacity,
		capacity:    capacity,
		signal:      make(chan struct{}, capacity),
	}

	for i := 0; i < int(capacity); i++ {
		cap.signal <- struct{}{}
	}

	return cap
}

func (w *weightedCap) Consume(ctx context.Context, n int64) (release func(), err error) {
	if n > w.maxCapacity {
		return func() {}, &ExceedingCapacityErr{n, w.maxCapacity}
	}
	if atomic.LoadInt64(&w.capacity) >= n {
		w.signalMu.Lock()
		defer w.signalMu.Unlock()

		<-w.signal
		w.consume(ctx, n)
		return func() {
			atomic.AddInt64(&w.capacity, n)
			w.signal <- struct{}{}
		}, nil
	}

	w.signalMu.Lock()
	defer w.signalMu.Unlock()

	for atomic.LoadInt64(&w.capacity) < n {
		err := w.waitForSignalLocked(ctx)
		if err != nil {
			return func() {}, err
		}
	}

	w.consume(ctx, n)
	return func() {
		atomic.AddInt64(&w.capacity, n)
		w.signal <- struct{}{}
	}, nil
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

type ExceedingCapacityErr struct {
	Capacity    int64
	MaxCapacity int64
}

func (e *ExceedingCapacityErr) Error() string {
	return fmt.Sprintf("weightedcap: requested capacity (%d) higher than the maximum capacity (%d)", e.Capacity, e.MaxCapacity)
}
