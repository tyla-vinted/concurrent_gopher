package lesson2

import "sync"

// Lesson 2: channels, unbuffered and buffered.
// Use the function you work on: make run produce, make check produce, make hint produce

// Kevin set up the new production line on Monday. It does not work yet.
// Fix the functions in order: Produce, PackAll, then Line, which uses both.

// Produce builds the orders one after another with build, and sends each toy to out.
//
// Now: the packer is not told that all toys are finished, so it waits at the belt forever.
// TODO 1: the packer should know when the work is done.
func Produce(orders []Order, build func(Order) Toy, out chan<- Toy) {
	for _, o := range orders {
		out <- build(o)
	}
}

// PackAll packs every toy from in with pack until in is closed, and returns the boxes.
//
// Now: the packer goes home without a look at the belt.
// TODO 2: the packer keeps packing until nothing more will come. https://go.dev/tour/concurrency/4
func PackAll(in <-chan Toy, pack func(Toy) Box) []Box {
	var boxes []Box
	return boxes
}

// Line connects a producer and a packer with a belt: a channel that holds beltSize toys.
// build and pack do the work.
//
// Now: Kevin forgot that the producer and the packer must work at the same time.
// TODO 3: let the two work side by side.
//
// Now: Kevin also bought the cheapest belt: it ignores beltSize.
// TODO 4: replace Kevin's belt with a sized one.
func Line(orders []Order, beltSize int, build func(Order) Toy, pack func(Toy) Box) []Box {
	belt := make(chan Toy)
	Produce(orders, build, belt)
	return PackAll(belt, pack)
}

// TODO 5: compare belts. In the scene, every toy takes 500ms to build and 1.5s to pack, and a batch has 6 orders.
//
//	make run line BELT=0
//	make run line BELT=1
//	make run line BELT=3
//	make run line BELT=10
//
// 1. The line takes the same time with every belt. Why?
// 2. What changes for the producer when the belt gets bigger?
// 3. A bigger belt does not make the line faster. When is it still useful?

// Stretch: BuildLimited builds all orders with build at the same time, but never more than limit at once.
//
// Now: the safety inspector came by: the factory can only run [limit] workbenches at once,
// and right now every gopher grabs one.
// TODO: keep the factory within the inspector's rule. https://go-patterns.dev/sync/semaphore
func BuildLimited(orders []Order, limit int, build func(Order) Toy) []Toy {
	toys := make([]Toy, len(orders))
	var wg sync.WaitGroup
	for i, o := range orders {
		wg.Go(func() {
			toys[i] = build(o)
		})
	}
	wg.Wait()
	return toys
}
