// Command factory shows the scene of one lesson in the terminal.
//
//	go run ./cmd/factory -belt 3 line    only the part for Line
//	go run ./cmd/factory 2               all parts of lesson 2
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"concurrent_gopher/internal/catalog"
	"concurrent_gopher/internal/scenes"
)

func main() {
	belt := flag.Int("belt", 0, "lesson 2: how many toys the belt holds")
	workers := flag.Int("workers", 3, "lesson 5: how many workers in the pool")
	flag.Parse()

	f, err := catalog.Parse("run", flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if f.Part == "" {
		fmt.Printf("%s has no scene: a data race does not show on the floor. Test it with: make check %s\n", f.Name, strings.ToLower(f.Name))
		return
	}
	os.Exit(scenes.Run(f.Lesson, scenes.Options{Func: f.Part, Belt: *belt, Workers: *workers}))
}
