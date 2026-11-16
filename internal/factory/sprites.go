package factory

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	spriteW = 14
	cardW   = 24
	cardH   = 10 // 5 sprite lines, 4 text lines, 1 blank
)

// The gopher from basic_gopher, in one pose for each state. Each pose has
// two frames, 5 lines each; the screen switches frames to animate it.
var (
	poseHammer = [2][5]string{{
		`  n____n   _ `,
		` ( O  O ) | |`,
		` (  ''  )_/  `,
		` (______)    `,
		`   d  b   [#]`,
	}, {
		`  n____n     `,
		` ( O  O )    `,
		` (  ''  )\   `,
		` (______) \= `,
		`   d  b  *[#]*`,
	}}
	poseDoze = [2][5]string{{
		`  n____n   z `,
		` ( -  - )  Z `,
		` (  ''  )    `,
		` (______)    `,
		`   d  b      `,
	}, {
		`  n____n  Z  `,
		` ( -  - )   z`,
		` (  ''  )    `,
		` (______)    `,
		`   d  b      `,
	}}
	poseSend = [2][5]string{{
		`  n____n     `,
		` ( O  O )  ? `,
		` (  ''  )=[#]`,
		` (______)    `,
		`   d  b      `,
	}, {
		`  n____n     `,
		` ( O  O )   ?`,
		` (  ''  )=[#]`,
		` (______)    `,
		`   d  b      `,
	}}
	poseReceive = [2][5]string{{
		`  n____n     `,
		` ( O  O )  ? `,
		` (  ''  )=[ ]`,
		` (______)    `,
		`   d  b      `,
	}, {
		`  n____n     `,
		` ( O  O )   ?`,
		` (  ''  )=[ ]`,
		` (______)    `,
		`   d  b      `,
	}}
	poseArmsCrossed = [2][5]string{{
		`  n____n     `,
		` ( O  O ) .. `,
		` (>-''-<)    `,
		` (______)    `,
		`   d  b      `,
	}, {
		`  n____n     `,
		` ( O  O ) ...`,
		` (>-''-<)    `,
		` (______)    `,
		`   d   b     `,
	}}
	poseDoor = [2][5]string{{
		`  n____n  |#|`,
		` ( O  O ) |#|`,
		` (  ''  )=|#|`,
		` (______) |#|`,
		`   d  b   |#|`,
	}, {
		`  n____n  |#|`,
		` ( O  O ) |#|`,
		` (  ''  )-|#|`,
		` (______) |#|`,
		`   d  b   |#|`,
	}}
	poseStuck = [2][5]string{{
		`  n____n     `,
		` ( X  X )    `,
		` (  ''  )    `,
		` (______)    `,
		`   d  b      `,
	}, {
		`  n____n     `,
		` ( X  X )    `,
		` (  ''  )    `,
		` (______)    `,
		`   d  b      `,
	}}
	poseStand = [2][5]string{{
		`  n____n     `,
		` ( O  O )    `,
		` (  ''  )/   `,
		` (______)    `,
		`   d  b      `,
	}, {
		`  n____n     `,
		` ( O  O )    `,
		` (  ''  )/   `,
		` (______)    `,
		`   d  b      `,
	}}
)

// pose picks the drawing for what g does now.
func pose(g gopher) [2][5]string {
	switch g.kind() {
	case kGone:
		return poseStand
	case kWorking:
		if g.item != "" {
			return poseHammer
		}
		return poseStand
	case kWaiting:
		return poseDoze
	}
	switch {
	case strings.Contains(g.state, "nil chan"), g.state == "select (no cases)":
		return poseStuck
	case strings.HasPrefix(g.state, "chan send"):
		return poseSend
	case strings.HasPrefix(g.state, "chan receive"):
		return poseReceive
	case strings.Contains(g.state, "Mutex"):
		return poseDoor
	}
	return poseArmsCrossed
}

func kindColor(k kind) func(string) string {
	switch k {
	case kGone:
		return dim
	case kWorking:
		return green
	case kWaiting:
		return yellow
	}
	return red
}

// card draws one gopher: its pose, then who it is, where, and what it does.
func card(g gopher, frame int) []string {
	paint := kindColor(g.kind())
	p := pose(g)[(frame/2)%2]
	out := make([]string, 0, cardH)
	for _, l := range p {
		out = append(out, paint(pad(l, cardW)))
	}
	out = append(out,
		bold(pad(label(g), cardW)),
		pad(g.fn, cardW),
		dim(pad(g.where, cardW)),
		paint(pad(cardStatus(g), cardW)),
		pad("", cardW),
	)
	return out
}

func cardStatus(g gopher) string {
	switch g.kind() {
	case kWorking:
		if g.item != "" {
			return g.item
		}
		return g.state
	}
	return g.state
}

// label names a gopher after its goroutine id, as in a goroutine dump.
func label(g gopher) string { return fmt.Sprintf("Gopher #%d", g.id) }

// pad cuts or fills s to exactly n runes.
func pad(s string, n int) string {
	r := []rune(s)
	if len(r) > n-1 {
		return string(r[:n-1]) + " "
	}
	return s + strings.Repeat(" ", n-len(r))
}

// cards lays out the gophers in rows of cards that fit the width.
func cards(gs []gopher, frame, cols int) []string {
	var out []string
	for i := 0; i < len(gs); i += cols {
		row := gs[i:min(i+cols, len(gs))]
		cs := make([][]string, len(row))
		for k, g := range row {
			cs[k] = card(g, frame)
		}
		for l := range cardH {
			var b strings.Builder
			b.WriteString("  ")
			for _, c := range cs {
				b.WriteString(c[l])
			}
			out = append(out, strings.TrimRight(b.String(), " "))
		}
	}
	return out
}

var (
	termW, termH = 100, 40
	termChecked  time.Time
)

// termSize asks the terminal for its size, at most once a second.
func termSize() (w, h int) {
	if time.Since(termChecked) < time.Second {
		return termW, termH
	}
	termChecked = time.Now()
	if c, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && c > 0 {
		termW = c
	}
	if r, err := strconv.Atoi(os.Getenv("LINES")); err == nil && r > 0 {
		termH = r
	}
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return termW, termH
	}
	defer tty.Close()
	cmd := exec.Command("stty", "size")
	cmd.Stdin = tty
	out, err := cmd.Output()
	if err != nil {
		return termW, termH
	}
	f := strings.Fields(string(out))
	if len(f) == 2 {
		if r, err := strconv.Atoi(f[0]); err == nil && r > 0 {
			termH = r
		}
		if c, err := strconv.Atoi(f[1]); err == nil && c > 0 {
			termW = c
		}
	}
	return termW, termH
}
