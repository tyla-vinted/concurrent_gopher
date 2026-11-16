package lesson2

import (
	"testing"
	"testing/synctest"
	"time"
)

// Inside synctest.Test the clock is fake: time.Sleep returns at once, and
// time.Since tells how long the code would take for real. When every
// goroutine in the test waits for another one, synctest stops the test with
// "deadlock".

func orders(toys ...string) []Order {
	out := make([]Order, len(toys))
	for i, t := range toys {
		out[i] = Order{ID: i + 1, Toy: t}
	}
	return out
}

func checkBoxes(t *testing.T, got []Box, os []Order) {
	t.Helper()
	if len(got) != len(os) {
		t.Fatalf("got %d boxes for %d orders", len(got), len(os))
	}
	for i, o := range os {
		if want := (Toy{OrderID: o.ID, Name: o.Toy}); got[i].Toy != want {
			t.Errorf("box %d holds %v; want %v", i, got[i].Toy, want)
		}
	}
}

func TestProduce(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		out := make(chan Toy, 10) // room for all toys, so Produce never waits
		Produce(orders("robot", "kite"), Build, out)
		for _, want := range []Toy{{1, "robot"}, {2, "kite"}} {
			if got := <-out; got != want {
				t.Fatalf("got %v; want %v", got, want)
			}
		}
		select {
		case _, ok := <-out:
			if ok {
				t.Fatal("Produce sent more toys than there are orders")
			}
		default:
			t.Fatal("out is not closed after Produce returns: close it when all toys are sent")
		}
	})
}

func TestPackAll(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		in := make(chan Toy, 3)
		in <- Toy{1, "robot"}
		in <- Toy{2, "kite"}
		in <- Toy{3, "kite"}
		close(in)
		checkBoxes(t, PackAll(in, Pack), orders("robot", "kite", "kite"))
	})
}

func TestLine(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		os := orders("robot", "kite", "yoyo", "drum")
		start := time.Now()
		boxes := Line(os, 0, Build, Pack)
		took := time.Since(start)
		checkBoxes(t, boxes, os)
		if took != 1300*time.Millisecond {
			t.Errorf("4 toys took %v; want 1.3s (100ms to build the first, then 4 × 300ms to pack)", took)
		}
	})
}

func TestLine_Buffered(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		os := orders("robot", "kite", "yoyo", "drum")
		start := time.Now()
		var lastBuilt time.Duration
		build := func(o Order) Toy {
			toy := Build(o)
			lastBuilt = time.Since(start)
			return toy
		}
		boxes := Line(os, 4, build, Pack)
		took := time.Since(start)
		checkBoxes(t, boxes, os)
		if took != 1300*time.Millisecond {
			t.Errorf("4 toys took %v; want 1.3s: a belt does not make the packer faster", took)
		}
		if lastBuilt != 400*time.Millisecond {
			t.Errorf("with a belt of 4, the producer built its last toy at %v; want 400ms: with room on the belt, it does not wait for the packer", lastBuilt)
		}
	})
}

func TestStretch_BuildLimited(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		os := orders("robot", "robot", "robot", "robot", "robot", "robot")
		start := time.Now()
		toys := BuildLimited(os, 3, Build)
		took := time.Since(start)
		for i, o := range os {
			if toys[i] != (Toy{o.ID, o.Toy}) {
				t.Errorf("toys[%d] = %v; want the toy for %v", i, toys[i], o)
			}
		}
		switch {
		case took < 200*time.Millisecond:
			t.Errorf("6 robots on 3 benches took %v; want 200ms: more than 3 were built at the same time", took)
		case took > 200*time.Millisecond:
			t.Errorf("6 robots on 3 benches took %v; want 200ms: are 3 built at the same time?", took)
		}
	})
}
