package lesson4

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

// Inside synctest.Test the clock is fake: time.Sleep and timers return at
// once, and time.Since tells how long the code would take for real.
// synctest also fails a test when goroutines still wait when it ends.

func orders(n int) []Order {
	out := make([]Order, n)
	for i := range out {
		out[i] = Order{ID: i + 1, Toy: "robot"}
	}
	return out
}

func TestTake(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		orders := make(chan Order, 1)
		orders <- Order{1, "robot"}
		o, err := Take(context.Background(), orders)
		if err != nil || o != (Order{1, "robot"}) {
			t.Errorf("Take = %v, %v; want order 1, nil", o, err)
		}
	})
}

func TestTake_Cancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(time.Second, cancel) // the Boss says stop after 1s
		orders := make(chan Order)          // no order comes
		start := time.Now()
		_, err := Take(ctx, orders)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Take = %v; want context.Canceled (from ctx.Err())", err)
		}
		if took := time.Since(start); took != time.Second {
			t.Errorf("Take with a cancel after 1s took %v; want 1s", took)
		}
	})
}

func TestTake_Deadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_, err := Take(ctx, make(chan Order)) // no order comes
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Take with a ctx that ends after 500ms = %v; want context.DeadlineExceeded", err)
		}
	})
}

func TestShift(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		toys, err := Shift(context.Background(), orders(3), Build, 2*time.Second)
		if err != nil || len(toys) != 3 {
			t.Errorf("Shift(3 orders, 2s) = %d toys, %v; want 3 toys, nil", len(toys), err)
		}
		if took := time.Since(start); took != 1200*time.Millisecond {
			t.Errorf("took %v; want 1.2s", took)
		}
	})
}

func TestShift_Timeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		toys, err := Shift(context.Background(), orders(10), Build, time.Second)
		took := time.Since(start)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Shift(10 orders, 1s): err = %v; want context.DeadlineExceeded", err)
		}
		if len(toys) != 2 {
			t.Errorf("Shift(10 orders, 1s) = %d toys; want 2 (at 400ms and 800ms)", len(toys))
		}
		if took != time.Second {
			t.Errorf("took %v; want 1s: the shift ends after length", took)
		}
	})
}

func TestShift_Cancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(500*time.Millisecond, cancel) // the boss ends the shift
		start := time.Now()
		_, err := Shift(ctx, orders(10), Build, time.Hour)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v; want context.Canceled", err)
		}
		if took := time.Since(start); took != 500*time.Millisecond {
			t.Errorf("took %v; want 500ms: stop when the caller cancels ctx", took)
		}
	})
}
