// Package catalog maps the functions that learners write to their lesson
// and to the part of the scene that shows them. It imports no lesson, so
// the tool still builds when a lesson does not compile.
package catalog

import (
	"fmt"
	"strconv"
	"strings"
)

// Func is one function that learners write.
type Func struct {
	Name   string
	Lesson int
	Part   string // the scene part that shows it, or "" when it has no scene
}

// Funcs lists them in lesson order.
var Funcs = []Func{
	{"BuildAll", 1, "BuildAll"},
	{"CountKinds", 1, ""}, // no scene: a data race does not show on the floor
	{"Count", 1, ""},      // no scene: a data race does not show on the floor
	{"Produce", 2, "Produce"},
	{"PackAll", 2, "PackAll"},
	{"Line", 2, "Line"},
	{"BuildLimited", 2, "BuildLimited"},
	{"Next", 3, "Next"},
	{"Fetch", 3, "Fetch"},
	{"Ship", 3, "Ship"},
	{"Fastest", 3, "Fastest"},
	{"Take", 4, "Take"},
	{"Shift", 4, "Shift"},
	{"Pool", 5, "Pool"},
	{"PoolGroup", 5, "PoolGroup"},
	{"Throttle", 5, "Throttle"},
}

// Find returns the function with this name, in any case.
func Find(name string) (Func, bool) {
	for _, f := range Funcs {
		if strings.EqualFold(f.Name, strings.TrimSpace(name)) {
			return f, true
		}
	}
	return Func{}, false
}

// Parse reads what the learner typed after the make command: a lesson
// number, or a function name in any case. For a lesson, f is empty.
func Parse(what string) (lesson int, f Func, err error) {
	what = strings.TrimSpace(what)
	if what == "" {
		return 0, Func{}, fmt.Errorf("say which function, e.g. make run ship\n%s", List())
	}
	if n, err := strconv.Atoi(what); err == nil {
		if n < 1 || n > 5 {
			return 0, Func{}, fmt.Errorf("there is no lesson %d: the lessons are 1 to 5", n)
		}
		return n, Func{}, nil
	}
	f, ok := Find(what)
	if !ok {
		return 0, Func{}, fmt.Errorf("there is no function %q in the lessons\n%s", what, List())
	}
	return f.Lesson, f, nil
}

// List names the functions of each lesson.
func List() string {
	var b strings.Builder
	b.WriteString("The functions:\n")
	for n := 1; n <= 5; n++ {
		var names []string
		for _, f := range Funcs {
			if f.Lesson == n {
				names = append(names, f.Name)
			}
		}
		fmt.Fprintf(&b, "  lesson %d: %s\n", n, strings.Join(names, ", "))
	}
	return b.String()
}
