// Package factory draws the factory floor in the terminal. It shows every
// goroutine that runs code from lessons/, and what that goroutine does now.
//
// The lessons call Work, Label and Say. In tests the screen is
// off: Work only sleeps, and the other calls do nothing.
package factory

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	on      atomic.Bool // true while the screen runs
	looping atomic.Bool // true in a terminal: scenes feed orders until q

	mu               sync.Mutex
	labels           = map[int64]string{} // goroutine id → the item it works on
	helpers          = map[int64]string{} // goroutine id → name of a factory hand
	failMsg          string               // set by Fail: why the scene stopped
	noDeadlockBefore time.Time            // set by Expect: something will still happen until then
	running          string               // the learner function the scene runs now
	logs             []string             // the last lines from Say
	results          []resultLine         // lines that stay until the scene ends
	phase            string
)

// Work shows that the calling goroutine works on item, and waits for d.
func Work(item string, d time.Duration) {
	defer Label(item)()
	time.Sleep(d)
}

// Label shows that the calling goroutine works on item, until it calls done.
func Label(item string) (done func()) {
	if !on.Load() {
		return func() {}
	}
	id := goid()
	mu.Lock()
	labels[id] = item
	mu.Unlock()
	return func() {
		mu.Lock()
		delete(labels, id)
		mu.Unlock()
	}
}

// Hand starts f in a goroutine that the floor shows as a gopher with the
// given name, although it runs no lesson code: a factory hand on the other
// end of a channel, so that a scene can show one learner function alone.
func Hand(name string, f func()) {
	go func() {
		if on.Load() {
			id := goid()
			mu.Lock()
			helpers[id] = name
			mu.Unlock()
			defer func() {
				mu.Lock()
				delete(helpers, id)
				mu.Unlock()
			}()
		}
		f()
	}()
}

// Say writes a line to the log under the floor.
func Say(format string, a ...any) {
	if !on.Load() {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	logs = append(logs, fmt.Sprintf(format, a...))
	if len(logs) > logLines {
		logs = logs[len(logs)-logLines:]
	}
}

type resultLine struct {
	key, text string
}

// Result writes a line that stays on the screen until the scene ends.
func Result(format string, a ...any) {
	mu.Lock()
	defer mu.Unlock()
	results = append(results, resultLine{"", fmt.Sprintf(format, a...)})
}

// Show writes the result line for key, in place of the last one for key:
// a scene that runs batch after batch keeps one line for each part.
func Show(key, format string, a ...any) {
	mu.Lock()
	defer mu.Unlock()
	text := fmt.Sprintf(format, a...)
	for i := range results {
		if results[i].key == key {
			results[i].text = text
			return
		}
	}
	results = append(results, resultLine{key, text})
}

// Phase starts a part of the scene. A new part clears the log; the next
// batch of the same part keeps it, so the screen does not jump.
func Phase(name string) {
	reportLeaks()
	mu.Lock()
	if name != phase {
		phase, logs = name, nil
	}
	mu.Unlock()
}

// Check reports goroutines that are still there after a batch (leaks).
func Check() { reportLeaks() }

// Looping reports whether the scene should feed batch after batch, until
// the learner quits. It is true in a terminal, after Run starts.
func Looping() bool { return looping.Load() }

// Expect tells the floor that something will still happen within d, e.g. a
// cancel from a timer that the floor cannot see. Until then, gophers that
// all wait are not a deadlock.
func Expect(d time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	noDeadlockBefore = time.Now().Add(d)
}

// Running tells the floor which learner function the scene runs now, so
// that the results can name it when it never finishes.
func Running(fn string) {
	mu.Lock()
	defer mu.Unlock()
	running = fn
}

// Fail stops the scene: the floor shows the reason in red, and the program
// exits, as after a deadlock. A factory hand calls it when it waited in
// vain for the learner's code.
func Fail(format string, a ...any) {
	mu.Lock()
	defer mu.Unlock()
	failMsg = fmt.Sprintf(format, a...)
}

// Stuck reports whether the scene leaked goroutines or failed: then it
// must stop feeding orders.
func Stuck() bool {
	mu.Lock()
	defer mu.Unlock()
	return anyLeak || failMsg != ""
}

const logLines = 6

// goid returns the id of the calling goroutine, as in a goroutine dump.
func goid() int64 {
	var buf [64]byte
	b := buf[:runtime.Stack(buf[:], false)]
	b = b[len("goroutine "):]
	for i, c := range b {
		if c == ' ' {
			b = b[:i]
			break
		}
	}
	id, _ := strconv.ParseInt(string(b), 10, 64)
	return id
}
