package scenes

import (
	"context"
	"time"

	"concurrent_gopher/internal/factory"
	"concurrent_gopher/lessons/lesson4"
)

// buildTime4 is how long a toy takes in the lesson 4 scenes: slow, so that
// each step can be seen. The tests use lesson4.BuildTime.
const buildTime4 = time.Second

// lesson4Orders returns batch number batch of n orders.
func lesson4Orders(batch, n int) []lesson4.Order {
	out := make([]lesson4.Order, n)
	for i := range out {
		id := batch*n + i + 1
		out[i] = lesson4.Order{ID: id, Toy: toyNames[(id-1)%len(toyNames)]}
	}
	return out
}

// build4 builds a toy in buildTime4. It stops early when ctx ends.
func build4(ctx context.Context, o lesson4.Order) (lesson4.Toy, error) {
	defer factory.Label("build " + o.String())()
	select {
	case <-time.After(buildTime4):
	case <-ctx.Done():
		return lesson4.Toy{}, ctx.Err()
	}
	factory.Say("built %v", o)
	return lesson4.Toy{OrderID: o.ID, Name: o.Toy}, nil
}

// bossCancels returns a ctx that the Boss cancels after d.
func bossCancels(d time.Duration) (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	t := time.AfterFunc(d, func() {
		factory.Say("the Boss called cancel()")
		cancel()
	})
	return ctx, func() { t.Stop(); cancel() }
}

// take shows one gopher that waits for an order that never comes. The Boss
// calls cancel() after 2s. With a correct Take, the gopher goes home then;
// Kevin's gopher waits forever, and the floor stops.
func take(batch int) {
	factory.Phase("Take: no order comes; the Boss calls cancel() after 2s")
	orders := make(chan lesson4.Order) // nobody sends
	ctx, stop := bossCancels(2 * time.Second)
	defer stop()
	factory.Expect(2 * time.Second) // the cancel is still coming: not a deadlock yet
	var err error
	took := timed(func() { _, err = lesson4.Take(ctx, orders) })
	factory.Show("Take", "Take: no order in %v, err: %v", took, err)
	factory.Say("stopped waiting: %v", err)
}

// shift shows a shift of 3.5s for 6 orders of 1s: a correct Shift builds 3,
// then the shift ends.
func shift(batch int) {
	factory.Phase("Shift of 3.5s for 6 orders × 1s")
	var toys []lesson4.Toy
	var err error
	took := timed(func() {
		toys, err = lesson4.Shift(context.Background(), lesson4Orders(batch, 6), build4, 3500*time.Millisecond)
	})
	factory.Show("Shift", "Shift: %d of 6 toys in %v, err: %s", realToys(toys), took, errText(err))
	if err != nil {
		factory.Say("the shift ended: %v", err)
	} else {
		factory.Say("built every order: the shift never ended")
	}
}

// realToys counts the toys that were built.
func realToys(toys []lesson4.Toy) int {
	n := 0
	for _, t := range toys {
		if t.OrderID != 0 {
			n++
		}
	}
	return n
}
