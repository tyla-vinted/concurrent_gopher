package lesson4

import (
	"context"
	"time"
)

// Lesson 4: context cancellation.
// Use the function you work on: make run take, make check take, make hint take

// The Boss has rules: a shift ends on time, and when the Boss says stop,
// everybody stops. Right now the gophers work all night.

// Take waits for the next order on [orders]. When [ctx] ends first, it returns ctx.Err().
//
// Now: Kevin's Take waits for an order, even when the Boss says stop.
// TODO 1: when the Boss says stop, stop. https://pkg.go.dev/context#Context
func Take(ctx context.Context, orders <-chan Order) (Order, error) {
	return <-orders, nil
}

// Shift builds the orders one after another with build, for a shift of the given length,
// and returns the toys finished in time.
// The error is nil when all orders are done, context.DeadlineExceeded when the shift ended first,
// or context.Canceled when the caller cancelled ctx.
//
// Now: the shift never ends and the union is not amused.
// TODO 2: the shift needs to define its end. https://pkg.go.dev/context#WithTimeout
func Shift(ctx context.Context, orders []Order, build func(context.Context, Order) (Toy, error), length time.Duration) ([]Toy, error) {
	var toys []Toy
	for _, o := range orders {
		t, err := build(ctx, o)
		if err != nil {
			return toys, err
		}
		toys = append(toys, t)
	}
	return toys, nil
}
