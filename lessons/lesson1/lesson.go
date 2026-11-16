package lesson1

// Lesson 1: goroutines.
// Use the function you work on: make run buildall, make check buildall, make hint buildall

// Welcome to the Gopher Factory. Toys are ordered all day, the Foreman
// counts every second, and Kevin, the intern, writes code that looks right.

// BuildAll calls work for every order and returns the toys.
//
// Now: one lonely gopher builds every order, one after another, and the customers are getting restless.
// TODO: hire more gophers, and don't hand over the toys before every one is built. https://pkg.go.dev/sync#WaitGroup
func BuildAll(orders []Order, work func(Order) Toy) []Toy {
	toys := make([]Toy, len(orders))
	for i, o := range orders {
		toys[i] = work(o)
	}
	return toys
}

// CountKinds builds every order with your BuildAll and counts the toys of each
// kind, e.g. {"robot": 3, "kite": 2}.
//
// Now: a colleague, Kevin, wrote CountKinds and left for a holiday.
// TODO: find the data race and fix it. https://pkg.go.dev/sync#Mutex
func CountKinds(orders []Order, work func(Order) Toy) map[string]int {
	counts := map[string]int{}
	BuildAll(orders, func(o Order) Toy {
		t := work(o)
		counts[t.Name]++
		return t
	})
	return counts
}

// Stretch: Count builds every order with your BuildAll, and returns how many
// toys it built. The Foreman wants one number at the end of the day.
//
// Now: the same colleague wrote this one too: n++ from many goroutines is a data race.
// TODO: fix it. A mutex works, but for one number there is something simpler: https://pkg.go.dev/sync/atomic#Int64
func Count(orders []Order, work func(Order) Toy) int {
	var n int
	BuildAll(orders, func(o Order) Toy {
		t := work(o)
		n++
		return t
	})
	return n
}
