package scenes

import (
	"fmt"
	"time"

	"concurrent_gopher/internal/factory"
	"concurrent_gopher/lessons/lesson1"
)

// lesson1Orders returns batch number batch of n orders.
func lesson1Orders(batch, n int) []lesson1.Order {
	out := make([]lesson1.Order, n)
	for i := range out {
		id := batch*n + i + 1
		out[i] = lesson1.Order{ID: id, Toy: toyNames[(id-1)%len(toyNames)]}
	}
	return out
}

func buildAll(batch int) {
	factory.Phase(fmt.Sprintf("BuildAll: 12 orders × %v", buildTime1))
	var toys []lesson1.Toy
	took := timed(func() { toys = lesson1.BuildAll(lesson1Orders(batch, 12), build1) })
	factory.Show("BuildAll", "BuildAll: %s in %v", plural(countToys(toys), "toy"), took)
}

// buildTime1 is how long a toy takes in the lesson 1 scenes: slow, so that
// each step can be seen, and as long as one batch on the floor. The tests
// use lesson1.Build (500ms).
const buildTime1 = time.Second

// build1 is Build, but buildTime1 long.
func build1(o lesson1.Order) lesson1.Toy {
	factory.Work("build "+o.String(), buildTime1)
	factory.Say("built %v", o)
	return lesson1.Toy{OrderID: o.ID, Name: o.Toy}
}

func countToys(toys []lesson1.Toy) int {
	n := 0
	for _, t := range toys {
		if t != (lesson1.Toy{}) {
			n++
		}
	}
	return n
}
