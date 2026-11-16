package lesson5

import (
	"context"
	"errors"
	"slices"
	"testing"
	"testing/synctest"
	"time"
)

// Inside synctest.Test the clock is fake: time.Sleep and timers return at
// once, and time.Since tells how long the code would take for real.
// synctest also fails a test when goroutines still wait when it ends.

func orders(toys ...string) []Order {
	out := make([]Order, len(toys))
	for i, t := range toys {
		out[i] = Order{ID: i + 1, Toy: t}
	}
	return out
}

// queue returns a closed channel that holds the orders.
func queue(os []Order) <-chan Order {
	ch := make(chan Order, len(os))
	for _, o := range os {
		ch <- o
	}
	close(ch)
	return ch
}

func robots(n int) []Order {
	out := make([]Order, n)
	for i := range out {
		out[i] = Order{ID: i + 1, Toy: "robot"}
	}
	return out
}

func TestPool(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		toys := Pool(queue(robots(12)), 4, Make)
		took := time.Since(start)
		var ids []int
		for _, toy := range toys {
			ids = append(ids, toy.OrderID)
		}
		slices.Sort(ids)
		if want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}; !slices.Equal(ids, want) {
			t.Errorf("got toys for orders %v; want %v", ids, want)
		}
		if took != 900*time.Millisecond {
			t.Errorf("12 orders × 300ms with 4 workers took %v; want 900ms", took)
		}
	})
}

func TestPoolGroup(t *testing.T) {
	t.Run("all good", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			toys, err := PoolGroup(context.Background(), queue(robots(12)), 4, Build)
			took := time.Since(start)
			if err != nil || len(toys) != 12 {
				t.Errorf("PoolGroup(12 orders, 4 workers) = %d toys, %v; want 12 toys, nil", len(toys), err)
			}
			if took != 900*time.Millisecond {
				t.Errorf("12 orders × 300ms with 4 workers took %v; want 900ms", took)
			}
		})
	})
	t.Run("a lemon", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			os := robots(12)
			os[5].Toy = "lemon"
			start := time.Now()
			_, err := PoolGroup(context.Background(), queue(os), 4, Build)
			if !errors.Is(err, ErrBroken) {
				t.Errorf("err = %v; want ErrBroken", err)
			}
			if took := time.Since(start); took != 450*time.Millisecond {
				t.Errorf("took %v; want 450ms: stop all builds when one fails", took)
			}
		})
	})
	t.Run("the Boss cancels", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			time.AfterFunc(400*time.Millisecond, cancel)
			start := time.Now()
			_, err := PoolGroup(ctx, queue(robots(12)), 4, Build)
			if !errors.Is(err, context.Canceled) {
				t.Errorf("err = %v; want context.Canceled", err)
			}
			if took := time.Since(start); took != 400*time.Millisecond {
				t.Errorf("took %v; want 400ms: stop when the caller cancels ctx", took)
			}
		})
	})
}

func TestStretch_Throttle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		start := time.Now()
		var at []time.Duration
		for range Throttle(ctx, queue(robots(4)), 100*time.Millisecond) {
			at = append(at, time.Since(start))
		}
		want := []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 300 * time.Millisecond, 400 * time.Millisecond}
		if !slices.Equal(at, want) {
			t.Errorf("orders came at %v; want %v", at, want)
		}

		ctx, cancel := context.WithCancel(ctx)
		ch := Throttle(ctx, queue(robots(10)), time.Hour)
		cancel()
		synctest.Wait()
		if _, ok := <-ch; ok {
			t.Error("after cancel, Throttle still sends orders")
		}
	})
}
