package lesson2

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

func (t Toy) String() string { return fmt.Sprintf("#%02d %s", t.OrderID, t.Name) }

// Box is a packed toy, ready to ship.
type Box struct {
	Toy Toy
}

// Every toy takes the same time, whatever it is.
const (
	BuildTime = 100 * time.Millisecond
	PackTime  = 300 * time.Millisecond
)

// Build makes the toy for o in BuildTime.
func Build(o Order) Toy {
	factory.Work("build "+o.String(), BuildTime)
	return Toy{OrderID: o.ID, Name: o.Toy}
}

// Pack puts t in a box in PackTime.
func Pack(t Toy) Box {
	factory.Work("pack "+t.String(), PackTime)
	factory.Say("packed %v", t)
	return Box{Toy: t}
}
