package lesson5

import (
	"context"
	"time"
)

// Lesson 5: worker pools and errgroup.
// Use the function you work on: make run pool, make check pool, make hint pool
//
// It is the holiday rush, and one gopher is not enough. The pool:
//
//	         ┌── worker ──┐
//	jobs ────┼── worker ──┼──── results ──> Pool collects
//	         └── worker ──┘
//	fan-out: the workers      fan-in: all workers send
//	share one jobs channel    to one results channel

// Pool builds all jobs with build, in a pool of workers, and returns the toys in any order.
//
// Now: one worker for the holiday rush is not enough: Pool builds every order itself.
// TODO 1: hire workers gophers for the rush, and bring all their toys back to one place.
func Pool(jobs <-chan Order, workers int, build func(Order) Toy) []Toy {
	var toys []Toy
	for o := range jobs {
		toys = append(toys, build(o))
	}
	return toys
}

// PoolGroup does the same as Pool, with errgroup.
// Handle errors during work and wait for all of them. When build fails, PoolGroup returns
// that error; when ctx ends, it returns ctx's error.
//
// Now: PoolGroup builds nothing.
// TODO 2: the same pool, with errgroup: one lemon stops the whole line, and the Boss hears why.
// https://pkg.go.dev/golang.org/x/sync/errgroup
func PoolGroup(ctx context.Context, jobs <-chan Order, workers int, build func(context.Context, Order) (Toy, error)) ([]Toy, error) {
	return nil, nil
}

// Stretch: Throttle passes the orders from in to its channel, but at most one every every.
// It closes its channel when in is closed, or when ctx ends.
//
// Now: Throttle passes the orders through with no limit, and the parts supplier sends angry letters.
// TODO: keep the parts supplier happy. https://pkg.go.dev/time#NewTicker
func Throttle(ctx context.Context, in <-chan Order, every time.Duration) <-chan Order {
	return in
}
