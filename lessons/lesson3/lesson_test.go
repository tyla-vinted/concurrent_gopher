package lesson3

import (
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

// Inside synctest.Test the clock is fake: time.Sleep and time.After return
// at once, and time.Since tells how long the code would take for real.
// When every goroutine in the test waits for another one, synctest stops
// the test with "deadlock".

func TestNext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		kites := make(chan Toy, 1)
		kites <- Toy{1, "kite"} // a kite is ready
		robots := make(chan Toy, 1)
		go func() {
			time.Sleep(time.Second)
			robots <- Toy{1, "robot"} // a robot comes after 1s
		}()

		start := time.Now()
		if got := Next(robots, kites); got.Name != "kite" {
			t.Fatalf("first Next = %v; want the kite, which was ready (the robot comes after 1s)", got)
		}
		if took := time.Since(start); took != 0 {
			t.Errorf("first Next took %v; want 0s: take the kite at once", took)
		}
		if got := Next(robots, kites); got.Name != "robot" {
			t.Errorf("second Next = %v; want the robot", got)
		}
		if took := time.Since(start); took != time.Second {
			t.Errorf("both took %v; want 1s", took)
		}
	})
}

func TestFetch_InTime(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		p, err := Fetch(Supplier("Acme", 200*time.Millisecond), 500*time.Millisecond)
		if err != nil || p != (Part{"bolt", "Acme"}) {
			t.Errorf("Fetch = %v, %v; want the bolt from Acme, nil", p, err)
		}
		if took := time.Since(start); took != 200*time.Millisecond {
			t.Errorf("took %v; want 200ms: return as soon as the part comes", took)
		}
	})
}

func TestFetch_TooSlow(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		_, err := Fetch(Supplier("Slowpoke", 2*time.Second), 500*time.Millisecond)
		if !errors.Is(err, ErrTimeout) {
			t.Errorf("Fetch from a supplier that needs 2s, with wait 500ms: err = %v; want ErrTimeout", err)
		}
		if took := time.Since(start); took != 500*time.Millisecond {
			t.Errorf("took %v; want 500ms: give up after wait", took)
		}
		time.Sleep(2 * time.Second) // the supplier finishes
	})
}

func TestShip(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		truck := make(chan Box, 2) // room for 2 boxes, and nobody unloads
		for i, want := range []bool{true, true, false, false} {
			if got := Ship(truck, Box{ID: i + 1}); got != want {
				t.Errorf("Ship box %d to a truck with room for 2: got %v; want %v", i+1, got, want)
			}
		}
		if len(truck) != 2 {
			t.Errorf("the truck holds %d boxes; want 2", len(truck))
		}
	})
}

func TestStretch_Fastest(t *testing.T) {
	// synctest also fails when goroutines still wait when the test ends.
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		p := Fastest([]func() Part{
			Call("Slowpoke", 900*time.Millisecond),
			Call("Acme", 200*time.Millisecond),
			Call("Bolts&Co", 500*time.Millisecond),
		})
		if p.From != "Acme" {
			t.Errorf("Fastest = part from %s; want the part from Acme", p.From)
		}
		if took := time.Since(start); took != 200*time.Millisecond {
			t.Errorf("took %v; want 200ms", took)
		}
		time.Sleep(time.Second) // the slower suppliers finish, or wait forever
	})
}
