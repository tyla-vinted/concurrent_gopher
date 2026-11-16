package factory

import (
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"
)

const (
	frameEvery   = 100 * time.Millisecond
	deadlockWait = 1500 * time.Millisecond // all blocked this long: deadlock
	deadlockShow = 4 * time.Second         // then show it this long, and stop
	leakGrace    = 600 * time.Millisecond  // time to finish after a part ends
	emptyGrace   = 300 * time.Millisecond  // an empty floor shorter than this is not drawn
	failGrace    = 500 * time.Millisecond  // after a hand gives up, the scene has this long to return
	width        = 78
)

var (
	frame   int
	color   = os.Getenv("NO_COLOR") == ""
	leaked  = map[int64]bool{}
	anyLeak bool
)

func paint(code, s string) string {
	if !color {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func bold(s string) string   { return paint("1", s) }
func dim(s string) string    { return paint("2", s) }
func green(s string) string  { return paint("32", s) }
func yellow(s string) string { return paint("33", s) }
func red(s string) string    { return paint("31", s) }

// Run draws the floor while play runs, and returns the exit code: 0 when
// all is well, 1 when goroutines leaked, 2 on a deadlock.
//
// In a terminal, the scene feeds batch after batch of orders (Looping)
// until the learner presses q. After a leak the
// scene stops, and Run waits for q. A deadlock shows its banner for a
// moment, then Run returns. Without a terminal, the scene plays once.
func Run(title string, play func()) int {
	on.Store(true)
	keys, restore := keyboard()
	defer restore()
	interactive := keys != nil
	looping.Store(interactive)

	start := time.Now()
	finished := make(chan struct{})
	go func() {
		play()
		close(finished)
	}()

	fmt.Print("\x1b[2J")
	tick := time.NewTicker(frameEvery)
	defer tick.Stop()
	var (
		allBlockedSince time.Time
		ended           bool // play returned: after a leak, or once without a terminal
		code            int
		lastGs          []gopher  // the last floor with gophers on it
		lastGophers     time.Time // when that was
		shown           []gopher  // the gophers of the last frame drawn
		failSeen        time.Time // when a factory hand gave up
	)
	footer := func() string { return "q: quit" }

	for {
		select {
		case k := <-keys:
			if k == 'q' || k == 'Q' || k == 27 {
				draw(title, start, sample(), "", "")
				fmt.Println()
				return code
			}

		case <-finished:
			finished, ended = nil, true
			reportLeaks() // a leak also fails the scene, with its reason
			mu.Lock()
			msg, leak := failMsg, anyLeak
			mu.Unlock()
			if msg != "" {
				notFinished()
			}
			// The final screen: every gopher that was on the floor stays.
			// One that is still there keeps its state; one whose goroutine
			// has ended stands still, with how it ended.
			label := "went home"
			if msg != "" {
				label = "stopped"
			}
			draw(title, start, finalFloor(sample(), lastGs, label), msg, "")
			fmt.Println()
			switch {
			case leak:
				return 1
			case msg != "":
				return 2
			}
			return code

		case <-tick.C:
			mu.Lock()
			msg := failMsg
			mu.Unlock()
			if msg != "" && finished != nil && failSeen.IsZero() {
				failSeen = time.Now()
			}
			if !failSeen.IsZero() && finished != nil && time.Since(failSeen) > failGrace {
				// A factory hand gave up, and the learner's code is still
				// stuck: stop now, with the reason.
				notFinished()
				gs := sample()
				if len(gs) == 0 {
					gs = shown
				}
				draw(title, start, gs, msg, "")
				fmt.Println()
				return 2
			}
			gs := sample()
			// Between two batches the floor is empty for a moment. Keep
			// the last frame, unless the floor stays empty for a while.
			if len(gs) == 0 && !ended && time.Since(lastGophers) < emptyGrace {
				gs = lastGs
			} else if len(gs) > 0 {
				lastGs, lastGophers = gs, time.Now()
			}
			banner := ""
			// After play ended, the gophers left are leaks, not a deadlock.
			// Before an expected event, waiting is not a deadlock either.
			mu.Lock()
			expecting := time.Now().Before(noDeadlockBefore)
			mu.Unlock()
			if !ended && !expecting && deadlocked(gs) {
				if allBlockedSince.IsZero() {
					allBlockedSince = time.Now()
				}
				if d := time.Since(allBlockedSince); d > deadlockWait {
					banner, code = deadlockBanner, 2
					if d > deadlockWait+deadlockShow {
						notFinished()
						draw(title, start, gs, banner, "")
						fmt.Println()
						return code
					}
				}
			} else {
				allBlockedSince = time.Time{}
			}
			f := ""
			if interactive {
				f = footer()
			}
			draw(title, start, gs, banner, f)
			shown = gs
		}
	}
}

var deadlockBanner = "DEADLOCK"

func deadlocked(gs []gopher) bool {
	if len(gs) == 0 {
		return false
	}
	for _, g := range gs {
		if g.kind() != kBlocked {
			return false
		}
	}
	return true
}

// reportLeaks waits a moment, then adds a result line for each lesson
// goroutine that is still there: nobody will ever stop it.
func reportLeaks() {
	if !on.Load() {
		return
	}
	var left []gopher
	for deadline := time.Now().Add(leakGrace); ; {
		left = left[:0]
		for _, g := range sample() {
			if !leaked[g.id] {
				left = append(left, g)
			}
		}
		if len(left) == 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if len(left) == 0 {
		return
	}
	mu.Lock()
	anyLeak = true
	mu.Unlock()
	Result("%s", red("LEAK: "+gophers(len(left))+" still there"))
	for _, g := range left {
		leaked[g.id] = true
	}
	Fail("%s", leakReason(left))
}

// notFinished adds a result line for the function that never returned. A
// function that returned has its own result line already (also when it
// left a goroutine behind), and keeps it.
func notFinished() {
	mu.Lock()
	defer mu.Unlock()
	if running == "" {
		return
	}
	for _, r := range results {
		if r.key == running {
			return
		}
	}
	results = append(results, resultLine{running, red(running + ": did not finish")})
}

// leakReason says in plain words what the first leaked gopher waits for.
func leakReason(left []gopher) string {
	g := left[0]
	var why string
	switch {
	case strings.HasPrefix(g.state, "chan send"):
		why = g.fn + " waits to send, and nobody takes it."
	case strings.HasPrefix(g.state, "chan receive"):
		why = g.fn + " waits to receive, and nothing comes."
	case strings.HasPrefix(g.state, "select"):
		why = g.fn + " waits in a select, and nothing happens."
	default:
		why = g.fn + " still waits: " + g.state + "."
	}
	if len(left) > 1 {
		if len(left) == 2 {
			why += " (1 more gopher waits too.)"
		} else {
			why += fmt.Sprintf(" (%d more gophers wait too.)", len(left)-1)
		}
	}
	return why
}

func draw(title string, start time.Time, gs []gopher, banner, footer string) {
	mu.Lock()
	lg := append([]string(nil), logs...)
	rs := append([]resultLine(nil), results...)
	mu.Unlock()

	frame++
	tw, th := termSize()
	w := max(width, min(tw-1, 140))

	var top, bottom []string
	stats := fmt.Sprintf("%5.1fs · goroutines %d", time.Since(start).Seconds(), runtime.NumGoroutine())
	head := "GOPHER FACTORY · " + title
	top = append(top,
		bold(head)+strings.Repeat(" ", max(1, w-len([]rune(head))-len(stats)))+dim(stats),
		dim(strings.Repeat("─", w)))
	running := 0
	for _, g := range gs {
		if !g.gone {
			running++
		}
	}
	top = append(top, bold("GOPHERS")+dim(fmt.Sprintf("  %d running lesson code", running)))

	var results, log, end []string
	if len(rs) > 0 {
		results = append(results, "", bold("RESULTS"))
		for _, r := range rs {
			results = append(results, "  "+r.text)
		}
	}
	if len(lg) > 0 {
		log = append(log, "", bold("LOG"))
		for _, l := range lg {
			log = append(log, dim("  "+l))
		}
	}
	if banner != "" {
		end = append(end, "")
		for l := range strings.SplitSeq(banner, "\n") {
			end = append(end, red(bold(l)))
		}
	}
	if footer != "" {
		end = append(end, "", dim(footer))
	}

	// The floor is always cards. The results, banner and footer always
	// show; the log goes first when space is short. When the window is too
	// short for every card, show the rows that fit, and say how many are
	// hidden.
	// Keep room for the RESULTS block (a blank line, the heading, one line)
	// before the first result comes, so that the floor does not move then.
	room := th - 1 - len(top) - max(len(results), 3) - len(end)
	if room-len(log) >= cardH || len(gs) == 0 {
		room -= len(log)
	} else {
		log = nil
	}
	var floor []string
	cols := max(1, (w-2)/cardW)
	rows := (len(gs) + cols - 1) / cols
	switch {
	case len(gs) == 0:
		floor = []string{dim("  (none)")}
	case rows*cardH <= room:
		floor = cards(gs, frame, cols)
	default:
		fit := max(0, (room-1)/cardH) // one line for the note
		shown := min(len(gs), fit*cols)
		floor = append(cards(gs[:shown], frame, cols),
			yellow("  … "+gophers(len(gs)-shown)+" more"))
	}
	bottom = slices.Concat(results, log, end)

	var b strings.Builder
	for _, l := range slices.Concat(top, floor, bottom) {
		b.WriteString(l + "\x1b[K\n")
	}
	fmt.Print("\x1b[H" + b.String() + "\x1b[J")
}

// finalFloor merges the gophers that are still there (now) with those of the
// last frame that had gophers (last). A gopher that is gone is marked with
// label instead of its last state, so that the final screen does not show
// work that already stopped.
func finalFloor(now, last []gopher, label string) []gopher {
	alive := map[int64]gopher{}
	for _, g := range now {
		alive[g.id] = g
	}
	var out []gopher
	seen := map[int64]bool{}
	for _, g := range last {
		seen[g.id] = true
		if a, ok := alive[g.id]; ok {
			out = append(out, a)
			continue
		}
		g.gone, g.state, g.item = true, label, ""
		out = append(out, g)
	}
	for _, g := range now {
		if !seen[g.id] {
			out = append(out, g)
		}
	}
	return out
}

func gophers(n int) string {
	if n == 1 {
		return "1 gopher"
	}
	return fmt.Sprintf("%d gophers", n)
}
