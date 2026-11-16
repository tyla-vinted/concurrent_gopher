package lesson3

import (
	"fmt"
	"time"

	"concurrent_gopher/internal/factory"
)

// You do not need to change this file.

// Toy is a finished toy.
type Toy struct {
	ID   int
	Name string
}

func (t Toy) String() string { return fmt.Sprintf("#%02d %s", t.ID, t.Name) }

// Part is a part that a supplier delivers.
type Part struct {
	Name string
	From string
}

// Box is a packed toy, ready to ship.
type Box struct {
	ID int
}

// Supplier asks a supplier for a part, like a request that is sent now: it
// starts a goroutine that delivers the part after the given time, on the
// channel it returns. The channel has room for the part, so the goroutine
// never waits for a receiver, also when nobody waits for the part any more.
func Supplier(from string, after time.Duration) <-chan Part {
	out := make(chan Part, 1)
	go func() {
		factory.Work("deliver a bolt", after)
		out <- Part{Name: "bolt", From: from}
	}()
	return out
}

// Call returns a function that asks a supplier for a part and waits for the
// answer for d.
func Call(from string, d time.Duration) func() Part {
	return func() Part {
		factory.Work("call "+from, d)
		return Part{Name: "gear", From: from}
	}
}
