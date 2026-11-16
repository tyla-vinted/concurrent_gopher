package scenes

import (
	"context"
	"fmt"
	"time"

	"concurrent_gopher/internal/factory"
	"concurrent_gopher/lessons/lesson5"
)

// buildTime5 is how long a toy takes in the lesson 5 scenes: slow, so that
// each step can be seen. The tests use lesson5.BuildTime.
const buildTime5 = time.Second

func lesson5Orders(batch, n int) []lesson5.Order {
	out := make([]lesson5.Order, n)
	for i := range out {
		id := batch*n + i + 1
		out[i] = lesson5.Order{ID: id, Toy: toyNames[(id-1)%len(toyNames)]}
	}
	return out
}

// build5 builds a toy in buildTime5. A lemon fails after half that time.
func build5(ctx context.Context, o lesson5.Order) (lesson5.Toy, error) {
	defer factory.Label("build " + o.String())()
	d := buildTime5
	if o.Toy == "lemon" {
		d /= 2
	}
	select {
	case <-time.After(d):
	case <-ctx.Done():
		return lesson5.Toy{}, ctx.Err()
	}
	if o.Toy == "lemon" {
		factory.Say("failed %v: %v", o, lesson5.ErrBroken)
		return lesson5.Toy{}, fmt.Errorf("order %d: %w", o.ID, lesson5.ErrBroken)
	}
	factory.Say("built %v", o)
	return lesson5.Toy{OrderID: o.ID, Name: o.Toy}, nil
}

// jobs5 returns the orders on a closed channel.
func jobs5(orders []lesson5.Order) <-chan lesson5.Order {
	ch := make(chan lesson5.Order, len(orders))
	for _, o := range orders {
		ch <- o
	}
	close(ch)
	return ch
}

// make5 makes a toy in buildTime5. It never fails.
func make5(o lesson5.Order) lesson5.Toy {
	factory.Work("build "+o.String(), buildTime5)
	factory.Say("built %v", o)
	return lesson5.Toy{OrderID: o.ID, Name: o.Toy}
}

// pool shows the fan-out and fan-in: 6 orders, workers gophers.
func pool(batch, workers int) {
	orders := lesson5Orders(batch, 6)
	factory.Phase(fmt.Sprintf("Pool of %d workers, 6 orders × 1s (change it: make run pool W=2)", workers))
	var toys []lesson5.Toy
	took := timed(func() { toys = lesson5.Pool(jobs5(orders), workers, make5) })
	factory.Show("Pool", "Pool: %d of 6 toys in %v, %d workers", len(toys), took, workers)
}

// poolGroup shows PoolGroup with a lemon: order 5 fails, and errgroup stops
// every worker.
func poolGroup(batch, workers int) {
	orders := lesson5Orders(batch, 6)
	orders[4].Toy = "lemon"
	factory.Phase(fmt.Sprintf("PoolGroup with errgroup, limit %d, 6 orders × 1s; order %d is a lemon", workers, orders[4].ID))
	var toys []lesson5.Toy
	var err error
	took := timed(func() { toys, err = lesson5.PoolGroup(context.Background(), jobs5(orders), workers, build5) })
	factory.Show("PoolGroup", "PoolGroup: %d of 6 toys in %v, err: %s", len(toys), took, errText(err))
}

func throttle(batch int) {
	factory.Phase("Stretch: Throttle 4 orders to one every 1s")
	in := jobs5(lesson5Orders(batch, 4))
	n := 0
	took := timed(func() {
		for o := range lesson5.Throttle(context.Background(), in, time.Second) {
			n++
			factory.Say("passed %v", o)
		}
	})
	factory.Show("Throttle", "Throttle: %d orders in %v", n, took)
}
