package scenes

import (
	"fmt"
	"time"

	"concurrent_gopher/internal/factory"
	"concurrent_gopher/lessons/lesson2"
)

// The lesson 2 scenes build and pack slowly, so that each step can be
// seen. The tests use lesson2.Build and lesson2.Pack (100ms and 300ms).
const (
	buildTime2 = 500 * time.Millisecond
	packTime2  = 1500 * time.Millisecond
)

func build2(o lesson2.Order) lesson2.Toy {
	factory.Work("build "+o.String(), buildTime2)
	return lesson2.Toy{OrderID: o.ID, Name: o.Toy}
}

func pack2(t lesson2.Toy) lesson2.Box {
	factory.Work("pack "+t.String(), packTime2)
	factory.Say("packed %v", t)
	return lesson2.Box{Toy: t}
}

// lineOrders returns batch number batch of 6 orders.
func lineOrders(batch int) []lesson2.Order {
	const n = 6
	orders := make([]lesson2.Order, n)
	for i := range orders {
		id := batch*n + i + 1
		orders[i] = lesson2.Order{ID: id, Toy: toyNames[(id-1)%len(toyNames)]}
	}
	return orders
}

// produce shows Produce alone: a factory hand packs what it sends.
func produce(batch int) {
	factory.Phase("Produce: you build and send, a factory hand packs")
	orders := lineOrders(batch)
	belt := make(chan lesson2.Toy)
	done := make(chan struct{})
	factory.Hand("Packer", func() {
		defer close(done)
		packed := 0
		for {
			select {
			case t, ok := <-belt:
				if !ok {
					return
				}
				pack2(t)
				packed++
			case <-time.After(handWait):
				factory.Fail("The packer waited %v after toy %d: nothing came, and the belt is still open.", handWait, packed)
				return
			}
		}
	})
	took := timed(func() { lesson2.Produce(orders, build2, belt) })
	<-done
	factory.Show("Produce", "Produce: %s sent in %v", plural(len(orders), "toy"), took)
}

// packAll shows PackAll alone: a factory hand builds the toys and sends
// them, then closes the channel.
func packAll(batch int) {
	factory.Phase("PackAll: a factory hand builds and sends, you pack")
	orders := lineOrders(batch)
	belt := make(chan lesson2.Toy)
	done := make(chan struct{})
	factory.Hand("Producer", func() {
		defer close(done)
		defer close(belt)
		for i, o := range orders {
			t := build2(o)
			select {
			case belt <- t:
			case <-time.After(handWait):
				factory.Fail("The producer waited %v with toy %d, and nobody took it.", handWait, i+1)
				return
			}
		}
	})
	var boxes []lesson2.Box
	took := timed(func() { boxes = lesson2.PackAll(belt, pack2) })
	<-done
	factory.Show("PackAll", "PackAll: %s in %v", plural(len(boxes), "box"), took)
}

// line shows all three together.
func line(batch, belt int) {
	factory.Phase(fmt.Sprintf("Line with a belt of %d · every toy: build %v, pack %v", belt, buildTime2, packTime2))
	orders := lineOrders(batch)
	var boxes []lesson2.Box
	took := timed(func() { boxes = lesson2.Line(orders, belt, build2, pack2) })
	factory.Show("Line", "Line: %s in %v, belt of %d", plural(len(boxes), "box"), took, belt)
}

// buildLimited builds slowly (1s a toy), so that each step is easy to see:
// 3 build while 3 wait, then the waiting 3 take over.
func buildLimited(batch int) {
	const n, benches, each = 6, 3, time.Second
	factory.Phase(fmt.Sprintf("Stretch: BuildLimited, %d orders on %d workbenches, %v a toy", n, benches, each))
	orders := make([]lesson2.Order, n)
	for i := range orders {
		orders[i] = lesson2.Order{ID: batch*n + i + 1, Toy: "train"}
	}
	slowBuild := func(o lesson2.Order) lesson2.Toy {
		factory.Work("build "+o.String(), each)
		factory.Say("built %v", o)
		return lesson2.Toy{OrderID: o.ID, Name: o.Toy}
	}
	took := timed(func() { lesson2.BuildLimited(orders, benches, slowBuild) })
	factory.Show("BuildLimited", "BuildLimited: %s in %v, %d benches", plural(n, "toy"), took, benches)
}
