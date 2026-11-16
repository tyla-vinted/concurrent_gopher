package lesson3

import (
	"errors"
	"time"
)

// Lesson 3: select.
// Use the function you work on: make run next, make check next, make hint next

// Kevin is back from holiday. The suppliers: Acme is quick, and Slowpoke is slow.

// Next takes one toy from whichever belt has one ready, robots or kites.
//
// Now: Kevin's Next always waits at the robot belt, even when a kite is ready.
// TODO: don't let the kites wait. https://go.dev/tour/concurrency/5
func Next(robots <-chan Toy, kites <-chan Toy) Toy {
	return <-robots
}

// ErrTimeout tells that a supplier did not deliver in time.
var ErrTimeout = errors.New("supplier is too slow")

// Fetch waits for one part from supplier, but no longer than [wait]. Then it gives up with ErrTimeout.
//
// Now: Fetch waits as long as the supplier needs, and Slowpoke takes its time.
// TODO: the Foreman won't wait longer than wait. https://pkg.go.dev/time#After
func Fetch(supplier <-chan Part, wait time.Duration) (Part, error) {
	return <-supplier, nil
}

// Ship puts [b] on the truck if the truck has room. It never waits: when the truck is full, it returns false at once.
//
// Now: Kevin's Ship waits until the truck has room. The truck is full and nobody unloads it,
// so the packer stands in the yard with the box forever.
// TODO: the packer has more boxes to pack: it can't idle in the yard.
func Ship(truck chan<- Box, b Box) bool {
	truck <- b
	return true
}

// Stretch: Fastest asks every supplier at the same time and returns the part that comes first.
//
// Now: it returns the right part, but make run fastest shows a LEAK: the slower suppliers call back,
// and nobody answers the phone.
// TODO: let the slower suppliers hang up. What if a supplier takes an hour? Lesson 4 answers.
func Fastest(suppliers []func() Part) Part {
	results := make(chan Part)
	for _, s := range suppliers {
		go func() { results <- s() }()
	}
	return <-results
}
