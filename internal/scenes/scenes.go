// Package scenes runs the lesson code on the factory floor, with sample
// orders, so that learners can watch what their goroutines do.
package scenes

import (
	"fmt"
	"os"
	"strings"
	"time"

	"concurrent_gopher/internal/factory"
)

// Options come from the make variables: make run line BELT=3.
type Options struct {
	Func    string // only the part for this lesson function, e.g. "Line"
	Belt    int    // lesson 2: how many toys the belt holds
	Workers int    // lesson 5: how many workers in the pool
}

// part shows one function of a lesson. play runs one batch of orders;
// batch counts from 0, so that order numbers can go on from batch to batch.
type part struct {
	fn   string // the function in lesson.go that the part runs
	play func(batch int)
}

// Run plays the scene of lesson n and returns the exit code.
func Run(n int, o Options) int {
	var title string
	var parts []part
	switch n {
	case 1:
		title = "Lesson 1: goroutines"
		parts = []part{{"BuildAll", buildAll}}
	case 2:
		title = "Lesson 2: channels"
		parts = []part{{"Produce", produce}, {"PackAll", packAll}, {"Line", func(b int) { line(b, o.Belt) }}, {"BuildLimited", buildLimited}}
	case 3:
		title = "Lesson 3: select"
		parts = []part{{"Next", next}, {"Fetch", fetch}, {"Ship", ship}, {"Fastest", fastest}}
	case 4:
		title = "Lesson 4: context"
		parts = []part{{"Take", take}, {"Shift", shift}}
	case 5:
		title = "Lesson 5: worker pools and errgroup"
		parts = []part{{"Pool", func(b int) { pool(b, o.Workers) }}, {"PoolGroup", func(b int) { poolGroup(b, o.Workers) }}, {"Throttle", throttle}}
	default:
		return 2
	}

	if o.Func != "" {
		var names []string
		var picked []part
		for _, p := range parts {
			names = append(names, p.fn)
			if strings.EqualFold(p.fn, o.Func) {
				picked = append(picked, p)
			}
		}
		if len(picked) == 0 {
			fmt.Fprintf(os.Stderr, "Lesson %d has no part for %q. Try one of: %s\n", n, o.Func, strings.Join(names, ", "))
			fmt.Fprintf(os.Stderr, "For example: make run %s\n", strings.ToLower(names[0]))
			return 2
		}
		parts = picked
		title += " · " + parts[0].fn
	}

	// Play the function once. The floor then exits, and keeps the last
	// state of the gophers on the screen.
	return factory.Run(title, func() {
		for _, p := range parts {
			factory.Running(p.fn)
			p.play(0)
			factory.Check()
			if factory.Stuck() {
				return
			}
		}
	})
}

var toyNames = []string{"robot", "kite", "yoyo", "drum", "train", "ball"}

// timed runs f and returns how long it took, rounded for the screen.
func timed(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start).Round(100 * time.Millisecond)
}

// handWait is how long a factory hand waits for the learner's function
// before it gives up and says why.
const handWait = 3 * time.Second

// itoa formats an order number for a label.
func itoa(n int) string { return fmt.Sprintf("%02d", n) }

// batchNote says which batch this is, when there is more than one.
func batchNote(batch int) string {
	if batch == 0 {
		return ""
	}
	return fmt.Sprintf(" (batch %d)", batch+1)
}

func plural(n int, one string) string {
	switch {
	case n == 1:
		return fmt.Sprintf("1 %s", one)
	case one == "box":
		return fmt.Sprintf("%d boxes", n)
	}
	return fmt.Sprintf("%d %ss", n, one)
}

// errText shows a nil error as "none".
func errText(err error) string {
	if err == nil {
		return "none"
	}
	return err.Error()
}
