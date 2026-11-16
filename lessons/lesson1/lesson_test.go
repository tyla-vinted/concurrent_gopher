package lesson1

import (
	"maps"
	"testing"
	"testing/synctest"
	"time"
)

// Inside synctest.Test the clock is fake: time.Sleep returns at once, and
// time.Since tells how long the code would take for real.

func orders(toys ...string) []Order {
	out := make([]Order, len(toys))
	for i, t := range toys {
		out[i] = Order{ID: i + 1, Toy: t}
	}
	return out
}

func TestBuildAll(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		os := orders("robot", "kite", "robot", "yoyo", "kite", "robot", "drum", "yoyo", "kite", "robot", "drum", "kite")
		start := time.Now()
		toys := BuildAll(os, Build)
		took := time.Since(start)

		if len(toys) != len(os) {
			t.Fatalf("got %d toys for %d orders", len(toys), len(os))
		}
		for i, o := range os {
			if toys[i] != (Toy{OrderID: o.ID, Name: o.Toy}) {
				t.Errorf("toys[%d] = %+v; want the toy for order %v", i, toys[i], o)
			}
		}
		if took != 500*time.Millisecond {
			t.Errorf("12 orders took %v; want 500ms: build them all at the same time, and wait for all of them", took)
		}
	})
}

func TestBuildAll_NoOrders(t *testing.T) {
	if toys := BuildAll(nil, Build); len(toys) != 0 {
		t.Errorf("BuildAll(nil) = %v; want no toys", toys)
	}
}

func TestCountKinds(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		got := CountKinds(orders("robot", "kite", "robot", "yoyo", "kite", "robot", "robot", "kite"), Build)
		want := map[string]int{"robot": 4, "kite": 3, "yoyo": 1}
		if !maps.Equal(got, want) {
			t.Errorf("CountKinds = %v; want %v", got, want)
		}
		if took := time.Since(start); took != 500*time.Millisecond {
			t.Errorf("8 orders took %v; want 500ms: CountKinds uses your BuildAll, so do BuildAll first", took)
		}
	})
}

func TestStretch_Count(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		if got := Count(orders("robot", "kite", "robot", "yoyo", "kite", "robot", "robot", "kite"), Build); got != 8 {
			t.Errorf("Count = %d; want 8", got)
		}
		if took := time.Since(start); took != 500*time.Millisecond {
			t.Errorf("8 orders took %v; want 500ms: Count uses your BuildAll, so do BuildAll first", took)
		}
	})
}
