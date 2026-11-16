package lesson1

import (
	"fmt"
	"time"

	"concurrent_gopher/internal/factory"
)

// You do not need to change this file.

// Order is one toy that a customer wants.
type Order struct {
	ID  int
	Toy string
}

func (o Order) String() string { return fmt.Sprintf("#%02d %s", o.ID, o.Toy) }

// Toy is a finished order.
type Toy struct {
	OrderID int
	Name    string
}

// Build makes the toy for o in 500ms. Most of that time it waits, like a
// call to a slow service: the goroutine sleeps and needs no CPU.
func Build(o Order) Toy {
	factory.Work("build "+o.String(), 500*time.Millisecond)
	return Toy{OrderID: o.ID, Name: o.Toy}
}
