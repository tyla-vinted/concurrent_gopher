package lesson5

import (
	"context"
	"errors"
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

// ErrBroken tells that a toy could not be built.
var ErrBroken = errors.New("broken part")

// BuildTime is how long Make and Build take for one toy.
const BuildTime = 300 * time.Millisecond

// Make makes the toy for o in BuildTime. It never fails. The tests and the
// scene pass it to Pool as build.
func Make(o Order) Toy {
	factory.Work("build "+o.String(), BuildTime)
	return Toy{OrderID: o.ID, Name: o.Toy}
}

// Build makes the toy for o in BuildTime. An order for a "lemon" fails
// with ErrBroken after half that time. Build stops early when ctx ends.
// The tests and the scene pass it to PoolGroup as build.
func Build(ctx context.Context, o Order) (Toy, error) {
	defer factory.Label("build " + o.String())()
	d := BuildTime
	if o.Toy == "lemon" {
		d /= 2
	}
	select {
	case <-time.After(d):
	case <-ctx.Done():
		return Toy{}, ctx.Err()
	}
	if o.Toy == "lemon" {
		factory.Say("order %d is a lemon: %v", o.ID, ErrBroken)
		return Toy{}, fmt.Errorf("order %d: %w", o.ID, ErrBroken)
	}
	return Toy{OrderID: o.ID, Name: o.Toy}, nil
}
