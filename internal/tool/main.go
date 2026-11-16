// Command tool runs the workshop commands behind the Makefile.
//
//	go run ./internal/tool check <func>   test the lesson of func, func first
//	go run ./internal/tool hint <func>    a hint for the first failing test of func
package main

import (
	"concurrent_gopher/internal/catalog"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	hintState = ".hint-state"
	stretch   = "TestStretch"
)

var titles = map[int]string{
	1: "goroutines",
	2: "channels, unbuffered and buffered",
	3: "select",
	4: "context cancellation",
	5: "worker pools and errgroup",
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "check":
		err = check(lessonArg())
	case "hint":
		err = hint(lessonArg())
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: tool check <func> | hint <func>")
	fmt.Fprintln(os.Stderr, "From make: make check ship")
	os.Exit(2)
}

// focus is the function the learner named, e.g. "Fetch", and sel is the
// same in lower case, to repeat it in the commands that we suggest.
var focus, sel string

// testsFunc reports whether test name tests function fn: TestFetch,
// TestFetch_TooSlow and TestStretch_Fetch all test Fetch.
func testsFunc(name, fn string) bool {
	rest := strings.TrimPrefix(strings.TrimPrefix(name, "Test"), "Stretch_")
	return rest == fn || strings.HasPrefix(rest, fn+"_")
}

// focused returns the results for the focus function.
func focused(rs []result) []result {
	var out []result
	for _, r := range rs {
		if testsFunc(r.name, focus) {
			out = append(out, r)
		}
	}
	return out
}

// watch suggests make run for the function that the learner named, when it
// has a scene.
func watch() string {
	if f, ok := catalog.Find(focus); ok && f.Part != "" {
		return " Watch it work: make run " + strings.ToLower(f.Name)
	}
	return ""
}

// lessonArg reads the function from the arguments, and returns its lesson.
func lessonArg() int {
	if len(os.Args) < 3 {
		usage()
	}
	f, err := catalog.Parse(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	focus, sel = f.Name, strings.ToLower(f.Name)
	return f.Lesson
}

// failing returns the other functions of lesson n with a failing test:
// main exercises and stretch exercises apart.
func failing(n int, rs []result) (main, stretchFuncs []string) {
	for _, f := range catalog.Funcs {
		if f.Lesson != n || f.Name == focus {
			continue
		}
		for _, r := range rs {
			if testsFunc(r.name, f.Name) && !r.ok {
				if r.isStretch() {
					stretchFuncs = append(stretchFuncs, f.Name)
				} else {
					main = append(main, f.Name)
				}
				break
			}
		}
	}
	return main, stretchFuncs
}

func pkg(n int) string { return fmt.Sprintf("./lessons/lesson%d", n) }

// result is the outcome of one test.
type result struct {
	name string
	ok   bool
	out  string
}

func (r result) isStretch() bool { return strings.HasPrefix(r.name, stretch) }

// runTests builds the tests of lesson n once, with the race detector, and
// runs each test on its own: a test that deadlocks stops only itself.
// On a build error, results is nil and buildOut holds the compiler output.
func runTests(n int) (results []result, buildOut string, err error) {
	tmp, err := os.MkdirTemp("", "lesson-test-*")
	if err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(tmp)
	bin := filepath.Join(tmp, "lesson.test")
	if out, err := exec.Command("go", "test", "-c", "-race", "-o", bin, pkg(n)).CombinedOutput(); err != nil {
		return nil, string(out), nil
	}
	list, err := exec.Command(bin, "-test.list", ".").Output()
	if err != nil {
		return nil, "", fmt.Errorf("list tests: %w", err)
	}
	for name := range strings.FieldsSeq(string(list)) {
		if !strings.HasPrefix(name, "Test") {
			continue
		}
		cmd := exec.Command(bin, "-test.run", "^"+name+"$", "-test.timeout", "20s")
		cmd.Dir = pkg(n)
		out, err := cmd.CombinedOutput()
		results = append(results, result{name, err == nil, string(out)})
	}
	return results, "", nil
}

// firstFailing returns the first failing test, main tests before stretch.
func firstFailing(rs []result) (result, bool) {
	for _, stretchToo := range []bool{false, true} {
		for _, r := range rs {
			if !r.ok && r.isStretch() == stretchToo {
				return r, true
			}
		}
	}
	return result{}, false
}

// why explains the failures that look scary in the test output.
func why(r result) string {
	switch {
	case strings.Contains(r.out, "panic: test timed out"):
		return r.name + " never finished: a loop never ends, or a goroutine waits forever."
	case strings.Contains(r.out, "main bubble goroutine has exited but blocked goroutines remain"):
		return r.name + " returned, but left goroutines that wait forever: a goroutine leak."
	case strings.Contains(r.out, "deadlock: all goroutines in bubble are blocked"):
		return r.name + " got stuck: every goroutine waits for another one. That is a deadlock."
	case strings.Contains(r.out, "WARNING: DATA RACE"):
		return r.name + " has a data race: two goroutines use the same memory at the same time."
	}
	return ""
}

func check(n int) error {
	fmt.Printf("Lesson %d: %s\n\n", n, titles[n])
	rs, buildOut, err := runTests(n)
	if err != nil {
		return err
	}
	if rs == nil {
		fmt.Print(buildOut)
		return fmt.Errorf("\nYour code does not compile yet. Fix the error above, then run make check %s again", sel)
	}
	for _, r := range rs {
		mark := "✓"
		if !r.ok {
			mark = "✗"
		}
		note := ""
		if r.isStretch() {
			note = "   (stretch: if you have time)"
		}
		line := fmt.Sprintf("  %s %s%s", mark, r.name, note)
		if !testsFunc(r.name, focus) {
			line = dimText(line)
		}
		fmt.Println(line)
	}
	if r, failed := firstFailing(focused(rs)); failed {
		return explain(n, r)
	}
	fmt.Printf("\n%s passes its tests.%s\n", focus, watch())
	main, stretchFuncs := failing(n, rs)
	switch {
	case len(main) > 0:
		fmt.Printf("Other functions of lesson %d still fail: %s.\n", n, strings.Join(main, ", "))
	case len(stretchFuncs) > 0:
		fmt.Printf("Lesson %d is done. Time left? Try the stretch exercise: make check %s\n", n, strings.ToLower(stretchFuncs[0]))
	default:
		fmt.Printf("Lesson %d is done.\n", n)
	}
	return nil
}

// explain prints the output of a failing test, and what to do next.
func explain(n int, r result) error {
	fmt.Printf("\n── %s ──\n", r.name)
	if w := why(r); w != "" {
		fmt.Printf("%s\n\n", w)
	}
	fmt.Print(trimOutput(r.out))
	return fmt.Errorf("\nLesson %d is not done yet. Fix %s first (make hint %s helps), then run make check %s again", n, r.name, sel, sel)
}

func dimText(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\x1b[2m" + s + "\x1b[0m"
}

// trimOutput drops the last FAIL line and keeps at most 60 lines.
func trimOutput(out string) string {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "FAIL" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > 60 {
		lines = append(lines[:60], fmt.Sprintf("... (%d more lines)", len(lines)-60))
	}
	return strings.Join(lines, "\n") + "\n"
}

func hint(n int) error {
	rs, buildOut, err := runTests(n)
	if err != nil {
		return err
	}
	if rs == nil {
		_ = buildOut
		return fmt.Errorf("Your code does not compile yet. Run make check %s to see the error, and fix that first", sel)
	}
	r, failed := firstFailing(focused(rs))
	if !failed {
		fmt.Printf("%s passes its tests.%s\n", focus, watch())
		return nil
	}
	name := r.name
	if w := why(r); w != "" {
		fmt.Printf("%s\n\n", w)
	}

	hints := readHints(filepath.Join(".hints", fmt.Sprintf("lesson%d.txt", n)), name)
	if len(hints) == 0 {
		fmt.Printf("%s fails, and there is no hint for it. Read the test, or ask the host.\n", name)
		return nil
	}
	key := fmt.Sprintf("lesson%d/%s", n, name)
	seen := readState()
	i := min(seen[key], len(hints)-1)
	fmt.Printf("Hint %d of %d for %s:\n\n  %s\n\n", i+1, len(hints), name, strings.Replace(hints[i], "\n", "\n\n", 1))
	if i+1 < len(hints) {
		fmt.Printf("Still stuck? Run make hint %s again for the next hint.\n", sel)
	}
	seen[key] = i + 1
	return writeState(seen)
}

// readHints returns the hints under "## <name>" in a hints file. Each
// line is one hint. Lines that start with a tab are code: they belong to the
// hint above them, and keep their indentation.
func readHints(path, name string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var hints []string
	in := false
	for line := range strings.SplitSeq(string(data), "\n") {
		if code, ok := strings.CutPrefix(line, "\t"); ok && in && len(hints) > 0 {
			hints[len(hints)-1] += "\n    " + code
			continue
		}
		line = strings.TrimSpace(line)
		if h, ok := strings.CutPrefix(line, "## "); ok {
			in = strings.TrimSpace(h) == name
			continue
		}
		if in && line != "" {
			hints = append(hints, line)
		}
	}
	return hints
}

func readState() map[string]int {
	seen := map[string]int{}
	data, _ := os.ReadFile(hintState)
	for line := range strings.SplitSeq(string(data), "\n") {
		key, n, ok := strings.Cut(line, " ")
		if v, err := strconv.Atoi(n); ok && err == nil {
			seen[key] = v
		}
	}
	return seen
}

func writeState(seen map[string]int) error {
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%s %d\n", k, seen[k])
	}
	return os.WriteFile(hintState, []byte(b.String()), 0o644)
}
