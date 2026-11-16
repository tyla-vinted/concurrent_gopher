package factory

import (
	"cmp"
	"maps"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const lessonsPkg = "concurrent_gopher/lessons/"

// gopher is one goroutine that runs code from lessons/.
type gopher struct {
	id    int64
	fn    string // the lesson function, e.g. "Line.func1"
	where string // e.g. "lesson.go:31"
	state string // as in a goroutine dump, e.g. "chan send"
	item  string // what it works on, from Work or Label
	gone  bool   // final screen only: the goroutine has ended; state says how
}

// kind groups the states for the screen.
type kind int

const (
	kWorking kind = iota // sleeps in Work, or runs
	kWaiting             // select: it can wake up by itself (timer)
	kBlocked             // waits for another goroutine
	kGone                // final screen only: the goroutine has ended
)

func (g gopher) kind() kind {
	switch {
	case g.gone:
		return kGone
	case blocking(g.state):
		return kBlocked
	case g.item != "":
		return kWorking // it waits for a timer in Work or in a select
	case strings.HasPrefix(g.state, "select"):
		return kWaiting
	}
	return kWorking
}

// blocking reports whether a goroutine in this state waits for another
// goroutine, and cannot wake up by itself.
func blocking(state string) bool {
	return strings.HasPrefix(state, "chan ") ||
		strings.HasPrefix(state, "sync.") ||
		strings.HasPrefix(state, "semacquire") ||
		state == "select (no cases)"
}

// sample reads a goroutine dump and returns the lesson goroutines.
func sample() []gopher {
	buf := make([]byte, 1<<16)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}

	mu.Lock()
	lab := maps.Clone(labels)
	hands := maps.Clone(helpers)
	mu.Unlock()

	var out []gopher
	for block := range strings.SplitSeq(string(buf), "\n\n") {
		g, ok := parse(block)
		if name, isHand := hands[g.id]; isHand && g.id != 0 {
			g.fn, g.where, ok = name, "factory hand", true
		}
		if ok {
			g.item = lab[g.id]
			out = append(out, g)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].id < out[j].id })
	return out
}

// parse reads one goroutine from a dump. ok is false when the goroutine
// runs no lesson code.
func parse(block string) (g gopher, ok bool) {
	lines := strings.Split(strings.TrimSpace(block), "\n")
	head, ok := strings.CutPrefix(lines[0], "goroutine ")
	if !ok {
		return g, false
	}
	idText, rest, _ := strings.Cut(head, " [")
	g.id, _ = strconv.ParseInt(idText, 10, 64)
	g.state, _, _ = strings.Cut(strings.TrimSuffix(rest, "]:"), ",")

	// Frames come in pairs, innermost first: function, then file:line. The
	// "created by" pair at the end is the parent, not a frame of this
	// goroutine. The name is the outermost lesson frame: the function the
	// goroutine runs. The place is the innermost line in lesson.go, where
	// the learner's code is now, or else the innermost lesson frame.
	var inner, mine string
	for i := 1; i+1 < len(lines); i += 2 {
		fn := lines[i]
		if strings.HasPrefix(fn, "created by ") {
			break
		}
		where := shortFile(lines[i+1])
		if !strings.HasPrefix(fn, lessonsPkg) || strings.HasSuffix(where, "_test.go") {
			continue
		}
		g.fn = shortFunc(fn)
		if inner == "" {
			inner = where
		}
		if mine == "" && strings.HasPrefix(where, "lesson.go:") {
			mine = where
		}
	}
	if g.fn == "" {
		return g, false // g still has the id and state, for factory hands
	}
	g.where = cmp.Or(mine, inner)
	return g, true
}

// shortFunc turns "concurrent_gopher/lessons/lesson2.Line.func1(0x1, ...)"
// into "Line.func1".
func shortFunc(fn string) string {
	fn = strings.TrimPrefix(fn, lessonsPkg)
	if i := strings.LastIndex(fn, "("); i > 0 {
		fn = fn[:i]
	}
	fn = strings.ReplaceAll(fn, "[...]", "")
	if _, after, ok := strings.Cut(fn, "."); ok {
		fn = after
	}
	return fn
}

// shortFile turns "\t/a/b/lesson.go:31 +0x2c" into "lesson.go:31".
func shortFile(line string) string {
	line = strings.TrimSpace(line)
	if i := strings.LastIndex(line, " +0x"); i > 0 {
		line = line[:i]
	}
	return filepath.Base(line)
}
