package weightedcap_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/romainmenke/weightedcap"
)

func TestPushPop_NoLoad(t *testing.T) {
	cap := weightedcap.New(3)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
	defer cancel()

	// no load test at max capacity
	err := cap.Consume(ctx, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer cap.Release(3)
}

func TestPushPop_Load(t *testing.T) {
	cap := weightedcap.New(3)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	// add some load
	{
		err := cap.Consume(ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
	}

	// release load after some time on other goroutine
	go func() {
		time.Sleep(time.Millisecond * 5)
		cap.Release(2)
	}()

	// attempt to add more load
	{
		err := cap.Consume(ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		defer cap.Release(2)
	}
}

func TestPushPop_Timeout(t *testing.T) {
	cap := weightedcap.New(3)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()

	time.Sleep(time.Millisecond * 2)

	// plenty of capacity, so will not block and Push will succeed.
	{
		err := cap.Consume(ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
	}

	// not enough capacity and timeout happened, must return error.
	{
		err := cap.Consume(ctx, 2)
		if err != ctx.Err() {
			t.Fatal(fmt.Sprintf("expected ctx err, got : %v", err))
		}
	}
}

func TestPushPop_NegativeCap(t *testing.T) {
	cap := weightedcap.New(3)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancel()

	expectedErr := &weightedcap.ExceedingCapacityErr{5, 3}
	err := cap.Consume(ctx, 5)
	if err.Error() != expectedErr.Error() {
		t.Fatal(fmt.Sprintf("expected : %v, got : %v", expectedErr, err))
	}
}
