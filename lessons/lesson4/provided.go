package lesson4

import (
	"context"
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

// BuildTime is how long Build takes for one toy.
const BuildTime = 400 * time.Millisecond

// Build makes the toy for o in BuildTime. It stops early when ctx ends, and
// returns ctx.Err(). The tests and the scene pass it to Shift as build.
func Build(ctx context.Context, o Order) (Toy, error) {
	defer factory.Label("build " + o.String())()
	select {
	case <-time.After(BuildTime):
	case <-ctx.Done():
		return Toy{}, ctx.Err()
	}
	factory.Say("built %v", o)
	return Toy{OrderID: o.ID, Name: o.Toy}, nil
}
