package scenes

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"concurrent_gopher/internal/factory"
	"concurrent_gopher/lessons/lesson3"
)

// belt starts a factory hand that makes n toys of one kind, one every
// every, and sends each on the channel it returns. It gives up when nobody
// takes a toy for handWait. done is closed when the hand stops.
func belt(name, toy string, n int, every time.Duration, done *sync.WaitGroup) <-chan lesson3.Toy {
	ch := make(chan lesson3.Toy)
	done.Add(1)
	factory.Hand(name, func() {
		defer done.Done()
		for i := 1; i <= n; i++ {
			factory.Work(fmt.Sprintf("build #%02d %s", i, toy), every)
			select {
			case ch <- lesson3.Toy{ID: i, Name: toy}:
			case <-time.After(handWait):
				factory.Fail("%s waited %v with %s %d, and nobody took it.", name, handWait, toy, i)
				return
			}
		}
	})
	return ch
}

// next calls Next for every toy, and shows the order they come in.
func next(batch int) {
	factory.Phase("Next: robots come every 1s, kites every 1.5s; take whichever is ready")
	var hands sync.WaitGroup
	robots := belt("Robots", "robot", 4, time.Second, &hands)
	kites := belt("Kites", "kite", 3, 1500*time.Millisecond, &hands)
	var order []string
	took := timed(func() {
		for range 7 {
			t := lesson3.Next(robots, kites)
			order = append(order, t.Name)
			factory.Say("took %v", t)
		}
	})
	hands.Wait()
	factory.Show("Next", "Next: %d toys in %v: %s", len(order), took, strings.Join(order, ", "))
}

// fetch shows how many bolts Fetch got, and whom it gave up on.
func fetch(batch int) {
	const wait = 1500 * time.Millisecond
	factory.Phase("Fetch a bolt from Acme twice, then from Slowpoke; wait 1.5s for each")
	suppliers := []struct {
		name  string
		after time.Duration
	}{{"Acme", 500 * time.Millisecond}, {"Acme", 500 * time.Millisecond}, {"Slowpoke", 1700 * time.Millisecond}}
	got := 0
	var gaveUpOn []string
	start := time.Now()
	show := func() {
		line := fmt.Sprintf("Fetch: %d of %d bolts in %v", got, len(suppliers), time.Since(start).Round(100*time.Millisecond))
		if len(gaveUpOn) > 0 {
			line += ", gave up on " + strings.Join(gaveUpOn, ", ")
		}
		factory.Show("Fetch", "%s", line)
	}
	for _, s := range suppliers {
		if _, err := lesson3.Fetch(lesson3.Supplier(s.name, s.after), wait); err == nil {
			got++
			factory.Say("got a bolt from %s", s.name)
		} else {
			gaveUpOn = append(gaveUpOn, s.name)
			factory.Say("gave up on %s", s.name)
		}
		show()
	}
	// Slowpoke still delivers after Fetch gave up, to nobody: let it finish.
	time.Sleep(max(0, 2700*time.Millisecond-time.Since(start)))
}

// ship shows Ship with a truck that has room for 3 boxes, and nobody
// unloads it: a factory hand packs 5 boxes and ships each one.
func ship(batch int) {
	const room, boxes = 3, 5
	factory.Phase(fmt.Sprintf("Ship %d boxes to a truck with room for %d; nobody unloads it", boxes, room))
	truck := make(chan lesson3.Box, room)
	onTruck, onFloor := 0, 0
	start := time.Now()
	done := make(chan struct{})
	factory.Hand("Packer", func() {
		defer close(done)
		for i := 1; i <= boxes; i++ {
			id := batch*boxes + i
			factory.Work(fmt.Sprintf("pack box %d", id), 500*time.Millisecond)
			if lesson3.Ship(truck, lesson3.Box{ID: id}) {
				onTruck++
				factory.Say("put box %d on the truck", id)
			} else {
				onFloor++
				factory.Say("left box %d on the floor: the truck is full", id)
			}
		}
	})
	<-done
	if onTruck+onFloor == boxes {
		factory.Show("Ship", "Ship: %d of %d boxes on the truck in %v, %d on the floor",
			onTruck, boxes, time.Since(start).Round(100*time.Millisecond), onFloor)
	}
}

func fastest(batch int) {
	factory.Phase("Stretch: Fastest of 3 suppliers (1s, 2s, 3s)")
	start := time.Now()
	var p lesson3.Part
	took := timed(func() {
		p = lesson3.Fastest([]func() lesson3.Part{
			lesson3.Call("Slowpoke", 3*time.Second),
			lesson3.Call("Acme", time.Second),
			lesson3.Call("Bolts&Co", 2*time.Second),
		})
	})
	factory.Show("Fastest", "Fastest: a %s from %s in %v", p.Name, p.From, took)
	factory.Say("got a %s from %s", p.Name, p.From)
	time.Sleep(max(0, 3200*time.Millisecond-time.Since(start))) // the others finish, or leak
}
